package lib

import (
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ilivestrong/recipe-stats-calculator/config"
)

const (
	deliveryTimePattern = `(\d{1,2})(?:AM|PM)`
)

type (
	DeliveryTime       string
	uniqueRecipes      map[string]int
	recipesPerPostCode map[string][]Recipe

	// Domain of our app
	Recipe struct {
		Name         string       `json:"recipe"`
		Postcode     string       `json:"postcode"`
		DeliveryTime DeliveryTime `json:"delivery"`
	}
	// custom filters for evaluating certain results
	filter struct {
		PostCode          string
		DeliveryStartTime string
		DeliveryEndTime   string
		RecipeNames       []string
	}

	//count aggregation per recipe name
	CountPerRecipeInfo struct {
		RecipeName string `json:"recipe"`
		Count      int    `json:"count"`
	}
	// postcode with max deliveries
	BusiestPostcodeInfo struct {
		Postcode      string `json:"postcode"`
		DeliveryCount int    `json:"delivery_count"`
	}
	// delivery count aggregation per postcode in a time window
	CountPerPostCodeAndTimeInfo struct {
		Postcode      string `json:"postcode"`
		From          string `json:"from"`
		To            string `json:"to"`
		DeliveryCount int    `json:"delivery_count"`
	}
	// Eventual result of the calculation
	Result struct {
		UniqueRecipeCount       int                         `json:"unique_recipe_count"`
		CountPerRecipe          []CountPerRecipeInfo        `json:"count_per_recipe"`
		BusiestPostcode         BusiestPostcodeInfo         `json:"busiest_postcode"`
		CountPerPostCodeAndTime CountPerPostCodeAndTimeInfo `json:"count_per_postcode_and_time"`
		MatchByName             []string                    `json:"match_by_name"`
	}

	// Base aggregation result(s) from calculate() API
	// Further used in transform() API to compute additional results
	RecipeStats struct {
		mUniqueRecipes      uniqueRecipes
		mRecipesPerPostCode recipesPerPostCode
	}
	// CLI's primary APIs
	RecipeStatsCalculator interface {
		Calculate([]Recipe) []RecipeStats
		Transform(...RecipeStats) Result
	}
	// Concrete API client
	statsCalculator struct {
		RecipeStatsCalculator
		filter filter
	}
)

// Calculate handles the recipe stats calculation request.
// Based on number of recipes in the source, spin additional go routines to do expensive task.
// It returns a slice of RecipeStats which contains partial calculation results.
func (sc *statsCalculator) Calculate(recipes []Recipe) []RecipeStats {

	var wg sync.WaitGroup
	recipeThresholdPerWorker := len(recipes)

	// we only spin additional goroutines if we have atleast 2.5M records
	if len(recipes) > 2500000 {
		recipeThresholdPerWorker = 2500000
	}

	count := len(recipes) / recipeThresholdPerWorker
	rs := make(chan RecipeStats, count)

	// Launch workers per recipe count threshold
	for i := 0; i <= count-1; i++ {
		var recipeSet []Recipe
		if i == len(recipes)-1 {
			recipeSet = recipes[i*recipeThresholdPerWorker:]
		} else {
			recipeSet = recipes[i*recipeThresholdPerWorker : (i+1)*recipeThresholdPerWorker]
		}

		wg.Add(1)
		go processRecipeSet(i, recipeSet, &wg, rs)
	}

	// when work finishes, close the result channel
	go func() {
		wg.Wait()
		close(rs)
	}()

	processedRecipeStats := make([]RecipeStats, 0)
	for stat := range rs {
		processedRecipeStats = append(processedRecipeStats, stat)
	}

	return processedRecipeStats
}

// processRecipeSet acts as a worker and hence invoked as a go routine from Calculate().
// It primarily takes a slice of recipes, performs few aggregations on them.
// It then wraps the aggregations in a RecipeStats struct and returns as a channel.
func processRecipeSet(wokerID int, recipes []Recipe, wg *sync.WaitGroup, rs chan RecipeStats) {
	defer wg.Done()

	mUniqueRecipes := make(uniqueRecipes)
	mRecipesPerPostCode := make(recipesPerPostCode)

	for _, recipe := range recipes {
		mUniqueRecipes[recipe.Name]++ // increment each unique recipe count
		if existingRecipesPerPostCode, exist := mRecipesPerPostCode[recipe.Postcode]; !exist {
			mRecipesPerPostCode[recipe.Postcode] = []Recipe{recipe}
		} else {
			mRecipesPerPostCode[recipe.Postcode] = append(existingRecipesPerPostCode, recipe)
		}
	}

	// return aggregation as a channel value
	rs <- RecipeStats{
		mUniqueRecipes:      mUniqueRecipes,
		mRecipesPerPostCode: mRecipesPerPostCode,
	}
}

// Transform receives aggregation resultsets from distinct go routines.
// For each resultset it then merges the results, then derives additional results from them.
// It finally builds and returns a recipe stats calculation response as a Result struct.
func (sc *statsCalculator) Transform(recipeStatsList ...RecipeStats) Result {
	mUniqueRecipeCount := make(uniqueRecipes)
	mRecipesPerPostCode := make(recipesPerPostCode)
	busiestPostCode := BusiestPostcodeInfo{}

	for _, rs := range recipeStatsList {
		//merge unique recipe count from each recipe stats list
		for k, v := range rs.mUniqueRecipes {
			mUniqueRecipeCount[k] += v
		}
		// merge recipes per postcode from each recipe stats list
		// also identify busiest postcode
		for pc, recipesPerPC := range rs.mRecipesPerPostCode {
			mRecipesPerPostCode[pc] = append(mRecipesPerPostCode[pc], recipesPerPC...)
			recipeForPostCode, _ := mRecipesPerPostCode[pc]
			if len(recipeForPostCode) > busiestPostCode.DeliveryCount {
				busiestPostCode = BusiestPostcodeInfo{
					Postcode:      pc,
					DeliveryCount: len(recipesPerPC),
				}
			}
		}
	}

	CountPerRecipeResultList := make([]CountPerRecipeInfo, 0)
	for k, v := range mUniqueRecipeCount {
		CountPerRecipeResultList = append(CountPerRecipeResultList, CountPerRecipeInfo{
			RecipeName: k,
			Count:      v,
		})
	}
	sort.Slice(CountPerRecipeResultList, func(i, j int) bool {
		return CountPerRecipeResultList[i].RecipeName < CountPerRecipeResultList[j].RecipeName
	})

	var dc int
	if recipes, ok := mRecipesPerPostCode[sc.filter.PostCode]; ok {
		for _, recipe := range recipes {
			if recipe.DeliveryTime.isWithin(sc.filter.DeliveryStartTime, sc.filter.DeliveryEndTime) {
				dc++
			}
		}
	}

	filteredRecipeNames := make([]string, 0)
	for _, result := range CountPerRecipeResultList {
		for _, filter := range sc.filter.RecipeNames {
			if strings.Contains(strings.ToLower(result.RecipeName), strings.ToLower(filter)) {
				filteredRecipeNames = append(filteredRecipeNames, result.RecipeName)
			}
		}
	}

	return Result{
		UniqueRecipeCount: len(mUniqueRecipeCount),
		CountPerRecipe:    CountPerRecipeResultList,
		BusiestPostcode:   busiestPostCode,
		CountPerPostCodeAndTime: CountPerPostCodeAndTimeInfo{
			Postcode:      sc.filter.PostCode,
			From:          sc.filter.DeliveryStartTime,
			To:            sc.filter.DeliveryEndTime,
			DeliveryCount: dc,
		},
		MatchByName: filteredRecipeNames,
	}
}

func NewStatsCalculator(options *config.Options) RecipeStatsCalculator {
	return &statsCalculator{
		filter: filter{
			PostCode:          options.SearchByPostCode,
			DeliveryStartTime: options.DeliveryTimeWindow.StartTime,
			DeliveryEndTime:   options.DeliveryTimeWindow.EndTIme,
			RecipeNames:       options.SearchByRecipeNames,
		},
	}
}

// isWithin acts a utility func that helps to filter recipes based on user specified delivery start and end times.
func (dt DeliveryTime) isWithin(startTimeFilter, endTimeFilter string) bool {
	re := regexp.MustCompile(deliveryTimePattern)
	matches := re.FindAllStringSubmatch(string(dt), -1)

	startTime, err := parseTime(startTimeFilter)
	if err != nil {
		return false
	}

	endTime, err := parseTime(endTimeFilter)
	if err != nil {
		return false
	}

	rangeStart, err := parseTime(matches[0][0])
	if err != nil {
		return false
	}

	rangeEnd, err := parseTime(matches[1][0])
	if err != nil {
		return false
	}

	return !startTime.Before(rangeStart) && !endTime.After(rangeEnd)
}

// parseTime takes a string of format "1AM"/"10PM" representing a time and parses it into a time.Time object.
func parseTime(timeStr string) (time.Time, error) {
	AMPM := "AM"
	if strings.Contains(timeStr, "PM") {
		AMPM = "PM"
	}

	v := strings.Split(timeStr, AMPM)
	timeStr = v[0] + ":00" + AMPM

	layout := "3:04PM"
	t, _ := time.Parse(layout, timeStr)

	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, now.Location()), nil
}
