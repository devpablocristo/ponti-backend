package support

// func afipJwtData(c *gin.Context) (string, error) {
// 	// Obtener el token del contexto, ya validado por el middleware
// 	token, _ := c.Get("token")
// 	jwtToken := token.(*jwt.Token) // Ya sabemos que es *jwt.Token por el middleware

// 	// Extraer las claims del token
// 	claims := jwtToken.Claims.(jwt.MapClaims) // Cast seguro porque ya fue validado

// 	// Obtener el CUIL del token
// 	cuil, _ := claims["cuil"].(string) // Asumimos que el claim está presente porque el middleware ya lo validó

// 	// Si el CUIL está vacío, retornamos error (por si acaso)
// 	if cuil == "" {
// 		return "", errors.New("CUIL is missing in the token")
// 	}

// 	return cuil, nil
// }

// func getSecrets() (map[string]string, error) {
// 	// Crear un mapa para almacenar los secrets
// 	secrets := make(map[string]string)

// 	// Cargar los secrets cuando sea necesario
// 	afipSecret := viper.GetString("AFIP_CLIENT_SECRET")
// 	miArgSecret := viper.GetString("MIARG_CLIENT_SECRET")

// 	// Si los secrets están vacíos, retornamos error (por si acaso)
// 	if afipSecret == "" {
// 		return nil, errors.New("AFIP_CLIENT_SECRET is missing")
// 	}
// 	if miArgSecret == "" {
// 		return nil, errors.New("MIARG_CLIENT_SECRET is missing")
// 	}

// 	// Guardar los secretos en el mapa
// 	secrets["afip"] = afipSecret
// 	secrets["miarg"] = miArgSecret

// 	return secrets, nil
// }
