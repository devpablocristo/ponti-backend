package config

// Todas las variables de la aplicación
type AllConfigs struct {
	General    General    // Variables generales
	API        API        // configuración de la API
	HTTPServer HTTPServer // configuración del servidor HTTP
	Debugger   Debugger   // configuración del debugger
	DB         DB         // configuración de la base de datos
	Suggester  Suggester  // configuración del sugeridor
}
