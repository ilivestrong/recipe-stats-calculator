package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/ilivestrong/recipe-stats-calculator/config"
	"github.com/ilivestrong/recipe-stats-calculator/lib"

	env "github.com/joho/godotenv"
)

const (
	dataDir            = "data"
	defaultFixtureFile = "all.json"
	envFile            = ".env"
	ColorGreen         = "\033[32m"
	ColorYellow        = "\033[33m"
	ColorPurple        = "\033[35m"
	ColorCyan          = "\033[36m"
	ColorBlue          = "\033[34m"
	ColorWhite         = "\033[0m"
)

func main() {
	options := buildAppOptions()
	printMessage(ColorWhite, "Loading fixture...\n", time.Millisecond*10)
	recipes, err := loadRecipes(options.FixtureFilePath)
	if err != nil {
		log.Fatalf("Err: %v", err)
	}

	printMessage(ColorWhite, "\nCalculating stats, hang on...\n\n", time.Millisecond*10)
	sc := lib.NewStatsCalculator(options)
	stats := sc.Calculate(recipes)
	result := sc.Transform(stats...)
	printResult(result)
}

func loadRecipes(fixturefile string) (Recipes []lib.Recipe, err error) {
	jsonFile, err := os.Open(fixturefile)
	if err != nil {
		fmt.Fprint(os.Stderr, "failed to load fixture file. ", err)
		return nil, err
	}

	fileBytes, err := io.ReadAll(jsonFile)
	if err != nil {
		log.Println("failed to load recipes from source. ", err)
		return nil, err
	}

	if err := json.Unmarshal(fileBytes, &Recipes); err != nil {
		println("failed to load recipes from source. ", err)
		return nil, err
	}
	return Recipes, nil
}

func printResult(result lib.Result) {
	resp, _ := json.MarshalIndent(result, "", "    ")
	printMessage(ColorGreen, string(resp), time.Millisecond*1)
}

func defaultOptions() *config.Options {
	err := env.Load(envFile)
	fixtureFile, exist := os.LookupEnv("FIXTURE_FILE")
	if !exist || err != nil {
		println("no default/invalid fixture file found, loading default...")
		fixtureFile = "all.json"
	}

	return &config.Options{
		FixtureFilePath:     path.Join(dataDir, fixtureFile),
		SearchByRecipeNames: []string{"Potato", "Veggie", "Mushroom"},
		SearchByPostCode:    "10120",
		DeliveryTimeWindow: config.DeliveryTimeWindow{
			StartTime: "10AM",
			EndTIme:   "3PM",
		},
	}
}

func buildAppOptions() *config.Options {
	reader := bufio.NewReader(os.Stdin)
	customOptions := new(config.Options)
	defaultOpts := defaultOptions()
	input := ""
	customOptions.FixtureFilePath = defaultOpts.FixtureFilePath

	printWelcomeMessage()

	printMessage(ColorYellow, "Do you want to set custom options for the app ? Enter Y to accept else press <ENTER>...\n", time.Millisecond*10)
	input, _ = reader.ReadString('\n')
	if strings.TrimSpace(input) != "Y" {
		return defaultOpts
	}

	printMessage(ColorPurple, "1. Enter custom fixture file name \n", time.Millisecond*10)
	input, _ = reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if strings.TrimSpace(input) != "" {
		customOptions.FixtureFilePath = path.Join(dataDir, input)
	}
	printMessage(ColorBlue, "\n\n2. Enter recipe name(s) to search by. (Separate by spaces)\n", time.Millisecond*10)
	input, _ = reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if strings.TrimSpace(input) != "" {
		customOptions.SearchByRecipeNames = strings.Split(input, " ")
	}
	printMessage(ColorCyan, "\n\n3.1 Enter postcode for which to calculate maximum deliveries\n", time.Millisecond*10)
	input, _ = reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if strings.TrimSpace(input) != "" {
		customOptions.SearchByPostCode = input
	}
	printMessage(ColorCyan, "\n\n3.2 Enter the time window (startTime <SPACE> endTime) for which to count maximum deliveries by postcode entered above(if any)\n", time.Millisecond*10)
	input, _ = reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if strings.TrimSpace(input) != "" {
		times := strings.Split(input, " ")
		customOptions.DeliveryTimeWindow = config.DeliveryTimeWindow{StartTime: times[0], EndTIme: times[1]}
	}
	return customOptions

}

func printWelcomeMessage() {
	printMessage(ColorGreen, strings.Repeat("-", 80)+"\n", time.Millisecond*10)
	printMessage(ColorGreen, "\nWelcome to Recipe stats calculator app...\n\n", time.Millisecond*20)
	printMessage(ColorGreen, "Please specify the config options you may want to set. \n", time.Millisecond*20)
	printMessage(ColorYellow, "(If you don't wish to specify value for an option, just press <ENTER>...\n\n", time.Millisecond*30)
	printMessage(ColorGreen, strings.Repeat("-", 80)+"\n\n", time.Millisecond*10)
}

func printMessage(color string, message string, timeout time.Duration) {
	for idx := range message[:] {
		time.Sleep(timeout)
		fmt.Printf("%s%s", color, string(message[idx]))
	}
}
