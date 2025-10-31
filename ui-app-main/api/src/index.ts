import express from "express";
import dotenv from "dotenv";
import routes from "./routes";
import path from "path";

dotenv.config();

const app = express();
const PORT = process.env.PORT || 3000;

const frontendPath = path.join(__dirname, "public");
app.use(express.static(frontendPath));

app.use(express.json({ limit: "150mb" }));
app.use(express.urlencoded({ extended: true, limit: "150mb" }));

if (process.env.NODE_ENV === "development") {
  console.log("Dev mode!");

  const { server } = require("./mocks/server");
  server.listen();
}

app.use("/api", routes);

app.get("/*", (_, res) => {
  res.sendFile(path.join(frontendPath, "index.html"));
});

app.listen(PORT, () => {
  console.log(`Servidor corriendo en http://localhost:${PORT}`);
});
