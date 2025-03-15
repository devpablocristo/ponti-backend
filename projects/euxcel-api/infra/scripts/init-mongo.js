// init-mongo.js

// Seleccionar o crear la base de datos 'qh_db'
db = db.getSiblingDB('qh_db');

// Crear el usuario 'user' con la contraseña 'userpassword'
db.createUser({
  user: "user",
  pwd: "userpassword", // Cambia esto por una contraseña segura
  roles: [
    {
      role: "readWrite",
      db: "qh_db"
    }
  ]
});
