// init-mongo.js

// Seleccionar o crear la base de datos 'euxcel_api_db'
db = db.getSiblingDB('euxcel_api_db');

// Crear el usuario 'user' con la contraseña 'userpassword'
db.createUser({
  user: "user",
  pwd: "userpassword", // Cambia esto por una contraseña segura
  roles: [
    {
      role: "readWrite",
      db: "euxcel_api_db"
    }
  ]
});
