posible problema con mongo:

Para ejecutarlo manualmente, puedes usar el cliente de MongoDB. Por ejemplo, si tienes el script `init-mongo.js` y el contenedor está en ejecución, sigue estos pasos:

1. Abre una terminal y accede al contenedor (suponiendo que el contenedor se llame `mongodb`):

   ```bash
   docker exec -it mongodb mongo --username root --password rootpassword --authenticationDatabase admin
   ```

2. Una vez en la shell de Mongo, carga y ejecuta el script con el comando `load`. Si el script se encuentra en `/docker-entrypoint-initdb.d/init-mongo.js`, ejecútalo así:

   ```js
   load("/docker-entrypoint-initdb.d/init-mongo.js")
   ```

Esto ejecutará el script y creará el usuario en la base de datos especificada. Asegúrate de que la ruta del script sea correcta según dónde se encuentre dentro del contenedor.