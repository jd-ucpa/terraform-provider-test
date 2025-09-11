package test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// checkTimestampIsRecent vérifie que le timestamp est proche de l'heure actuelle (dans les 5 secondes)
func checkTimestampIsRecent(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		timestampStr := rs.Primary.Attributes["result"]
		if timestampStr == "" {
			return fmt.Errorf("Timestamp attribute is empty")
		}

		// Parser le timestamp
		timestamp, err := time.Parse("2006-01-02T15:04:05", timestampStr)
		if err != nil {
			return fmt.Errorf("Failed to parse timestamp '%s': %v", timestampStr, err)
		}

		// Vérifier que le timestamp est proche de l'heure actuelle (dans les 5 secondes)
		now := time.Now().UTC()
		diff := now.Sub(timestamp)
		if diff < 0 {
			diff = -diff
		}
		if diff > 5*time.Second {
			return fmt.Errorf("Timestamp '%s' is too far from current time '%s' (diff: %v)", 
				timestampStr, now.Format("2006-01-02T15:04:05"), diff)
		}

		return nil
	}
}

// checkTimestampWithOffset vérifie qu'un timestamp est correctement calculé avec un offset
func checkTimestampWithOffset(resourceName string, expectedOffset time.Duration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		timestampStr := rs.Primary.Attributes["result"]
		if timestampStr == "" {
			return fmt.Errorf("Timestamp attribute is empty")
		}

		// Parser le timestamp
		timestamp, err := time.Parse("2006-01-02T15:04:05", timestampStr)
		if err != nil {
			return fmt.Errorf("Failed to parse timestamp '%s': %v", timestampStr, err)
		}

		// Calculer le timestamp de référence (maintenant + offset)
		expectedTime := time.Now().UTC().Add(expectedOffset)
		
		// Vérifier que la différence est dans une marge acceptable (5 secondes)
		diff := expectedTime.Sub(timestamp)
		if diff < 0 {
			diff = -diff
		}
		if diff > 5*time.Second {
			return fmt.Errorf("Timestamp '%s' is not close to expected time '%s' (diff: %v)", 
				timestampStr, expectedTime.Format("2006-01-02T15:04:05"), diff)
		}

		return nil
	}
}

// checkTimestampWithTimezoneAndOffset vérifie qu'un timestamp est correctement calculé avec un timezone et un offset
func checkTimestampWithTimezoneAndOffset(resourceName string, timezone string, expectedOffset time.Duration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		timestampStr := rs.Primary.Attributes["result"]
		if timestampStr == "" {
			return fmt.Errorf("Timestamp attribute is empty")
		}

		// Charger le timezone
		loc, err := time.LoadLocation(timezone)
		if err != nil {
			return fmt.Errorf("Failed to load timezone '%s': %v", timezone, err)
		}

		// Parser le timestamp dans le timezone spécifié
		timestamp, err := time.ParseInLocation("2006-01-02T15:04:05", timestampStr, loc)
		if err != nil {
			return fmt.Errorf("Failed to parse timestamp '%s' in timezone '%s': %v", timestampStr, timezone, err)
		}

		// Calculer le timestamp de référence (maintenant dans le timezone + offset)
		expectedTime := time.Now().In(loc).Add(expectedOffset)
		
		// Vérifier que la différence est dans une marge acceptable (5 secondes)
		diff := expectedTime.Sub(timestamp)
		if diff < 0 {
			diff = -diff
		}
		if diff > 5*time.Second {
			return fmt.Errorf("Timestamp '%s' is not close to expected time '%s' in timezone '%s' (diff: %v)", 
				timestampStr, expectedTime.Format("2006-01-02T15:04:05"), timezone, diff)
		}

		return nil
	}
}

// TestAccTimestampDataSource_Basic teste le data source timestamp avec la configuration par défaut.
// Ce test vérifie que le data source retourne un timestamp valide au format yyyy-mm-ddThh:mm:ss.
func TestAccTimestampDataSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "test_timestamp" "test" {}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_timestamp.test", "id"),
					resource.TestCheckResourceAttrSet("data.test_timestamp.test", "result"),
					resource.TestCheckResourceAttrSet("data.test_timestamp.test", "id"),
					// Vérifier que le timestamp est au bon format (sans le 'Z' final)
					resource.TestMatchResourceAttr("data.test_timestamp.test", "result", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}$`)),
					// Vérifier que le timestamp est proche de l'heure actuelle
					checkTimestampIsRecent("data.test_timestamp.test"),
				),
			},
		},
	})
}

// TestAccTimestampDataSource_WithTimezone teste le data source timestamp avec un fuseau horaire spécifique.
// Ce test vérifie que le timestamp est retourné dans le bon fuseau horaire.
func TestAccTimestampDataSource_WithTimezone(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "test_timestamp" "test" {
						time_zone = "Europe/Paris"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_timestamp.test", "result"),
					resource.TestCheckResourceAttr("data.test_timestamp.test", "time_zone", "Europe/Paris"),
					// Vérifier que le timestamp est au bon format
					resource.TestMatchResourceAttr("data.test_timestamp.test", "result", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}$`)),
					// Vérifier que le timestamp est proche de l'heure actuelle dans le timezone Europe/Paris
					checkTimestampWithTimezoneAndOffset("data.test_timestamp.test", "Europe/Paris", 0),
				),
			},
		},
	})
}

// TestAccTimestampDataSource_WithTimeAdd teste le data source timestamp avec l'ajout de temps.
// Ce test vérifie que le timestamp est correctement modifié selon les paramètres time_add.
func TestAccTimestampDataSource_WithTimeAdd(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "test_timestamp" "test" {
						time_add {
							days = 1
							hours = 2
							minutes = 30
							seconds = 45
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_timestamp.test", "result"),
					// Vérifier que le timestamp est au bon format
					resource.TestMatchResourceAttr("data.test_timestamp.test", "result", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}$`)),
					// Vérifier que le timestamp est correctement calculé avec l'ajout de temps
					// 1 jour + 2 heures + 30 minutes + 45 secondes
					checkTimestampWithOffset("data.test_timestamp.test", 26*time.Hour + 30*time.Minute + 45*time.Second),
				),
			},
		},
	})
}

// TestAccTimestampDataSource_WithTimeSubtract teste le data source timestamp avec la soustraction de temps.
// Ce test vérifie que le timestamp est correctement modifié avec des valeurs négatives.
func TestAccTimestampDataSource_WithTimeSubtract(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "test_timestamp" "test" {
						time_add {
							days = -1
							hours = -2
							minutes = -30
							seconds = -45
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_timestamp.test", "result"),
					// Vérifier que le timestamp est au bon format
					resource.TestMatchResourceAttr("data.test_timestamp.test", "result", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}$`)),
					// Vérifier que le timestamp est correctement calculé avec la soustraction de temps
					// -1 jour - 2 heures - 30 minutes - 45 secondes
					checkTimestampWithOffset("data.test_timestamp.test", -(26*time.Hour + 30*time.Minute + 45*time.Second)),
				),
			},
		},
	})
}

// TestAccTimestampDataSource_WithTimezoneAndTimeAdd teste le data source timestamp avec
// un fuseau horaire et l'ajout de temps combinés.
func TestAccTimestampDataSource_WithTimezoneAndTimeAdd(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "test_timestamp" "test" {
						time_zone = "America/New_York"
						time_add {
							days = 2
							hours = -1
							minutes = 15
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_timestamp.test", "result"),
					resource.TestCheckResourceAttr("data.test_timestamp.test", "time_zone", "America/New_York"),
					// Vérifier que le timestamp est au bon format
					resource.TestMatchResourceAttr("data.test_timestamp.test", "result", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}$`)),
					// Vérifier que le timestamp est correctement calculé avec l'ajout de temps
					// 2 jours - 1 heure + 15 minutes = 47 heures + 15 minutes
					checkTimestampWithTimezoneAndOffset("data.test_timestamp.test", "America/New_York", 47*time.Hour + 15*time.Minute),
				),
			},
		},
	})
}

// TestAccTimestampDataSource_InvalidTimezone teste le data source timestamp avec un fuseau horaire invalide.
// Ce test vérifie que le data source retourne une erreur appropriée.
func TestAccTimestampDataSource_InvalidTimezone(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "test_timestamp" "test" {
						time_zone = "Invalid/Timezone"
					}
				`,
				ExpectError: regexp.MustCompile(`Invalid timezone`),
			},
		},
	})
}
