/* rpcs3-gameupdater main entry point */

package main

import (
	// TODO: these are the UI libs
	//"fyne.io/fyne/app"
	//"fyne.io/fyne/widget"
	// TODO: figure out if we really need this

	"bufio"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

/* parses the given config.yml file and returns the path to dev_hdd0 */

func getGamesPath(configYML string) string {
	printInfo("Parsing '" + configYML)
	path := ""
	file, err := os.Open(configYML)

	if err != nil {
		printError(fmt.Sprintf("Couldn't open '%s' (errorcode: %s)\n", configYML, err))
		return path
	}

	defer file.Close()

	emulatorDir := ""
	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadString('\n')
		if strings.Contains(line, "$(EmulatorDir):") {
			emulatorDir = strings.TrimSpace(strings.Split(line, ":")[1])
			if emulatorDir == "\"\"" {
				emulatorDir = filepath.Dir(configYML) + "/"
			}
			printDebug("emudir: " + emulatorDir)
		}
		if strings.Contains(line, "/dev_hdd0/") {
			path = strings.Replace(strings.TrimSpace(strings.Split(line, ":")[1]), "$(EmulatorDir)", emulatorDir, -1)
			printDebug("path: " + path)
		}

		// we use err to figure out end of input
		if err != nil {
			return path
		}
	}
}

/* replaces the variable in the URL with the gameID */

func getURLFromID(id string) string {
	return fmt.Sprintf(urlPattern, id, id)
}

/* gets the game's version */

func getCategoryAndVersion(path string) (string, float64) {
	var folder string
	if strings.Contains(path, "/disc/") {
		folder = "/PS3_GAME"
	}
	// finds the PARAM.SFO
	params, err := filepath.Glob(path + folder + "/PARAM.SFO")
	if err != nil {
		printError("Error finding %s/**/PARAM.sfo  (errorcode: %s)\n", path, err)
		return "", 0.0
	}
	param := params[0]
	file, err := os.Open(param)
	defer file.Close()

	if err != nil {
		printError(fmt.Sprintf("Couldn't open '%s' (errorcode: %s)\n", param, err))
		return "", 0.0
	}
	kvp := readParamSFO(file)
	cat := getCategory(kvp)
	ver := getAppVersion(kvp)
	// in case there is no app version, use version instead
	if ver == "" {
		ver = getVersion(kvp)
	}
	versionF, err := strconv.ParseFloat(ver[0:5], 64)
	if err != nil {
		printError("Couldn't convert '%s' (errorcode: '%s')\n", ver, err)
	}
	return cat, versionF
}

/* gets games URLs and versions from a specific folder */

func getGamesFromFolder(games map[string]*GameInfo, path string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		printError("Couldn't open '%s' (errorcode: '%s')\n", path, err)
		return
	}

	for _, file := range files {
		if file.IsDir() && file.Name() != "TEST12345" && file.Name() != ".locks" && !strings.Contains(file.Name(), "INST") {
			url := getURLFromID(file.Name())

			category, version := getCategoryAndVersion(path + file.Name())

			if game, ok := games[file.Name()]; ok {
				if game.Version < version {
					game.Version = version
				}
			} else {
				game := GameInfo{
					Category: category,
					URL:      url,
					Version:  version,
				}
				games[file.Name()] = &game
			}
		}
	}
}

/* gets games URLs and versions from the various folders */

func getGames(path string) map[string]*GameInfo {
	// first from the disc folder
	games := make(map[string]*GameInfo)
	getGamesFromFolder(games, path+"disc/")

	// then the game folder
	getGamesFromFolder(games, path+"game/")

	return games
}

func getGamesFromServer() {
	printInfo("downloading using config.yml")

	path := getGamesPath(fetchConfig().ConfigYMLPath)
	games := getGames(path)

	for gameID, game := range games {
		printDebug("gameID: %s, url: %s, version: %s", gameID, game.URL, game.Version)
		url := game.URL
		//printInfo("fetching URL: '%s'", url)

		// we need this because we can't verify the signature
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		httpClient := &http.Client{Transport: transport,
			Timeout: time.Duration(conf.DLTimeout) * time.Second}

		// TODO: retry logic goes here
		response, err := httpClient.Get(url)

		if err != nil {
			printError("Error: Can't open url '%s'", url)
		}
		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)

		if err != nil {
			printError("can't read response body.")
			break
		}
		patch := TitlePatch{}
		err = xml.Unmarshal([]byte(body), &patch)

		if err != nil {
			printError("can't parse response XML.")
			continue
		}

		for i := range patch.Tag.Package {
			printInfo("title '%s' (%s) version %s url '%s' SHA '%s':",
				patch.Tag.Package[i].Paramsfo.TITLE,
				patch.Titleid,
				patch.Tag.Package[i].Version,
				patch.Tag.Package[i].URL,
				patch.Tag.Package[i].SHA1)
			version, err := strconv.ParseFloat(patch.Tag.Package[i].Version, 64)
			if err != nil {
				printError("Couldn't convert '%s' (errorcode: '%s')\n", patch.Tag.Package[i].Version, err)
			}
			if version < game.Version {
				printDebug("Version %f is inferior to current of %f", version, game.Version)
				continue
			}
			//downloadFileWithRetries(conf.PkgDLPath, patch.Tag.Package[i].URL, patch.Tag.Package[i].SHA1)
		}

	}
}

func main() {
	parseArguments()
	initConfig()
	initDownloader()
	getGamesFromServer()

	// test
	fmt.Printf("Terminal: %v\n", isTTY())

	// TODO: UI stuff
	//app := app.New()

	//w := app.NewWindow("Hello")
	//w.SetContent(widget.NewLabel("Hello Fyne!"))

	//w.ShowAndRun()
}
