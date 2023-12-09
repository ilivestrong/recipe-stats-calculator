package lib

import (
	"testing"

	"github.com/ilivestrong/recipe-stats-calculator/config"
)

func TestCalculate(t *testing.T) {

	tests := []struct {
		name string
		arg  []Recipe
		want []RecipeStats
	}{
		{
			name: "Should Calculate recipe stats successfully",
			arg: []Recipe{
				{
					Name:         "Creamy Dill Chicken",
					Postcode:     "10224",
					DeliveryTime: "Wednesday 1AM - 7PM",
				},
				{
					Name:         "Steakhouse-Style New York Strip",
					Postcode:     "10139",
					DeliveryTime: "Friday 7AM - 7PM",
				},
				{
					Name:         "Steakhouse-Style New York Strip",
					Postcode:     "10136",
					DeliveryTime: "Friday 11AM - 2PM",
				},
				{
					Name:         "Cherry Balsamic Pork Chops",
					Postcode:     "10130",
					DeliveryTime: "Saturday 1AM - 8PM",
				},
				{
					Name:         "Melty Monterey Jack Burgers",
					Postcode:     "10130",
					DeliveryTime: "Friday 6AM - 9PM",
				},
			},
			want: []RecipeStats{
				{
					mUniqueRecipes: map[string]int{
						"Creamy Dill Chicken":             1,
						"Steakhouse-Style New York Strip": 2,
						"Cherry Balsamic Pork Chops":      1,
						"Melty Monterey Jack Burgers":     1,
					},
					mRecipesPerPostCode: map[string][]Recipe{
						"10224": {{
							Name:         "Creamy Dill Chicken",
							Postcode:     "10224",
							DeliveryTime: "Wednesday 1AM - 7PM",
						}},
						"10139": {{
							Name:         "Steakhouse-Style New York Strip",
							Postcode:     "10139",
							DeliveryTime: "Friday 7AM - 7PM",
						}},
						"10136": {{
							Name:         "Steakhouse-Style New York Strip",
							Postcode:     "10139",
							DeliveryTime: "Friday 11AM - 2PM",
						}},
						"10130": {
							{
								Name:         "Cherry Balsamic Pork Chops",
								Postcode:     "10130",
								DeliveryTime: "Saturday 1AM - 8PM",
							},
							{
								Name:         "Melty Monterey Jack Burgers",
								Postcode:     "10130",
								DeliveryTime: "Friday 6AM - 9PM",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		rsc := NewStatsCalculator(&config.Options{
			SearchByRecipeNames: []string{"Jack"},
			SearchByPostCode:    "10224",
			DeliveryTimeWindow: config.DeliveryTimeWindow{
				StartTime: "1AM",
				EndTIme:   "9PM",
			},
		})
		t.Run(tt.name, func(t *testing.T) {
			got := rsc.Calculate(tt.arg)
			for k, v := range got[0].mUniqueRecipes {
				if tt.want[0].mUniqueRecipes[k] != v {
					t.Errorf("got: %v, want: %v", got, tt.want)
				}
			}
		})
	}
}

func TestTransform(t *testing.T) {

	tests := []struct {
		name          string
		options       *config.Options
		calculateArgs []Recipe
		want          Result
	}{
		{
			name: "Should get correct transformed Result",
			options: &config.Options{
				SearchByRecipeNames: []string{"Jack"},
				SearchByPostCode:    "10224",
				DeliveryTimeWindow: config.DeliveryTimeWindow{
					StartTime: "1AM",
					EndTIme:   "7PM",
				},
			},
			calculateArgs: []Recipe{
				{
					Name:         "Creamy Dill Chicken",
					Postcode:     "10224",
					DeliveryTime: "Wednesday 1AM - 7PM",
				},
				{
					Name:         "Creamy Dill Chicken",
					Postcode:     "10224",
					DeliveryTime: "Wednesday 5PM - 7PM",
				},
				{
					Name:         "Creamy Dill Chicken",
					Postcode:     "10224",
					DeliveryTime: "Thursday 12AM - 7PM",
				},
				{
					Name:         "Steakhouse-Style New York Strip",
					Postcode:     "10139",
					DeliveryTime: "Friday 7AM - 7PM",
				},
				{
					Name:         "Steakhouse-Style New York Strip",
					Postcode:     "10136",
					DeliveryTime: "Friday 11AM - 2PM",
				},
				{
					Name:         "Cherry Balsamic Pork Chops",
					Postcode:     "10130",
					DeliveryTime: "Saturday 1AM - 8PM",
				},
				{
					Name:         "Melty Monterey Jack Burgers",
					Postcode:     "10130",
					DeliveryTime: "Friday 6AM - 9PM",
				},
				{
					Name:         "Spinach Artichoke Pasta Bake",
					Postcode:     "10130",
					DeliveryTime: "Monday 9AM - 4PM",
				},
				{
					Name:         "Chicken Sausage Pizzas",
					Postcode:     "10130",
					DeliveryTime: "Friday 9AM - 11PM",
				},
			},
			want: Result{
				UniqueRecipeCount: 6,
				CountPerRecipe: []CountPerRecipeInfo{
					{
						RecipeName: "Creamy Dill Chicken",
						Count:      3,
					},
					{
						RecipeName: "Steakhouse-Style New York Strip",
						Count:      2,
					},
					{
						RecipeName: "Cherry Balsamic Pork Chops",
						Count:      1,
					},
					{
						RecipeName: "Melty Monterey Jack Burgers",
						Count:      1,
					},
					{
						RecipeName: "Spinach Artichoke Pasta Bake",
						Count:      1,
					},
					{
						RecipeName: "Chicken Sausage Pizzas",
						Count:      1,
					},
				},
				BusiestPostcode: BusiestPostcodeInfo{
					Postcode:      "10130",
					DeliveryCount: 4,
				},
				CountPerPostCodeAndTime: CountPerPostCodeAndTimeInfo{
					Postcode:      "10224",
					From:          "1AM",
					To:            "7PM",
					DeliveryCount: 2,
				},
				MatchByName: []string{"Melty Monterey Jack Burgers"},
			},
		},
	}

	for _, tt := range tests {
		rsc := NewStatsCalculator(tt.options)
		t.Run(tt.name, func(t *testing.T) {
			recipeStats := rsc.Calculate(tt.calculateArgs)
			got := rsc.Transform(recipeStats...)

			if got.BusiestPostcode.Postcode != tt.want.BusiestPostcode.Postcode {
				t.Errorf("got1: %#v, want: %#v", got, tt.want)
			}

			if got.CountPerPostCodeAndTime.DeliveryCount != tt.want.CountPerPostCodeAndTime.DeliveryCount {
				t.Errorf("got2: %#v, want: %#v", got, tt.want)
			}

			if got.MatchByName[0] != tt.want.MatchByName[0] {
				t.Errorf("got3: %#v, want: %#v", got, tt.want)
			}
		})
	}
}
