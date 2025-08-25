package test

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// LoadTestEnv charge les variables d'environnement depuis le fichier test.env
// dans le même répertoire que le fichier de test appelant
func LoadTestEnv(t *testing.T) {
	// Obtenir le répertoire du fichier de test appelant
	_, filename, _, _ := runtime.Caller(1)
	testDir := filepath.Dir(filename)
	envFile := filepath.Join(testDir, "test.env")
	
	file, err := os.Open(envFile)
	if err != nil {
		t.Fatalf("Impossible d'ouvrir %s: %v", envFile, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Ignorer les lignes vides et les commentaires
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Parser les variables d'environnement
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				
				// Supprimer les guillemets si présents
				value = strings.Trim(value, `"'`)
				
				// Définir la variable d'environnement
				os.Setenv(key, value)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Erreur lors de la lecture de test.env: %v", err)
	}
}

// RequireEnvVar vérifie qu'une variable d'environnement est définie
// et fait échouer le test avec un message d'erreur si elle est manquante
func RequireEnvVar(t *testing.T, varName string) string {
	value := os.Getenv(varName)
	if value == "" {
		t.Fatalf("La variable %s doit être définie dans test.env", varName)
	}
	return value
}

// SetupTestEnv charge les variables d'environnement depuis test.env et valide
// les variables spécifiées. Chaque test peut déclarer les variables dont il a besoin.
// Usage: SetupTestEnv(t, "ROLE_ARN", "INSTANCE_ID")
func SetupTestEnv(t *testing.T, requiredVars ...string) {
	// Charger les variables d'environnement depuis test.env
	LoadTestEnv(t)
	
	// Valider les variables requises spécifiées par le test
	for _, varName := range requiredVars {
		RequireEnvVar(t, varName)
	}
}
