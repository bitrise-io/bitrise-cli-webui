package main

import (
	"encoding/json"

	"fmt"
	bitrise "github.com/bitrise-io/bitrise-cli/bitrise"
	bitriseModel "github.com/bitrise-io/bitrise-cli/models/models_1_0_0"
	"github.com/kokomo88/bitrise-cli-webui/models"
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

// WriteBytesToFile ...
func WriteBytesToFile(pth string, fileCont []byte) error {
	if pth == "" {
		fmt.Println("No path provided")
	}

	file, err := os.Create(pth)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Println("Failed to close file:", err)
		}
	}()

	if _, err := file.Write(fileCont); err != nil {
		return err
	}

	return nil
}

func saveConfigToFile(pth string, bitriseConf bitriseModel.BitriseDataModel) error {
	contBytes, err := generateYAML(bitriseConf)
	if err != nil {
		return err
	}
	if err := bitrise.WriteBytesToFile(pth, contBytes); err != nil {
		return err
	}

	log.Println()
	log.Println("=> Init success!")
	log.Println("File created at path:", pth)

	return nil
}

func generateYAML(v interface{}) ([]byte, error) {
	bytes, err := yaml.Marshal(v)
	if err != nil {
		return []byte{}, err
	}
	return bytes, nil
}

func saveConfig(data bitriseModel.BitriseDataModel) {
	//fmt.Println(data)
	err := saveConfigToFile("./test/bitrise.yml", data)
	printError("saveConfigToFile:", err)
}
func readYAMLToBytes() []byte {
	bitriseConfig, err := bitrise.ReadBitriseConfig("./test/bitriseSafe.yml")
	//fmt.Printf("%#v", bitriseConfig)
	printError("BitriseData ReadBitriseConfig:", err)
	err = bitriseConfig.Normalize()
	printError("BitriseData Normalize:", err)
	err = bitriseConfig.Validate()
	printError("BitriseData Validate:", err)
	//bitriseConfig.FillMissingDeafults()
	var message = models.InitMessage{}
	message.Msg = bitriseConfig
	message.Type = "init"
	m, err := json.Marshal(&message)
	printError("Json encoding:", err)
	return m
}
func printError(from string, err error) {
	if err != nil {
		fmt.Println(from, err.Error())
	}
}
