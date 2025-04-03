import express from "express";
import { expressAnalytics } from "../analytics.js";
import * as dotenv from "dotenv";
dotenv.config();

const apiKey = process.env.API_KEY;

const app = express();

app.use(expressAnalytics(apiKey));

app.get("/", (req, res) => {
	res.send({ message: "Hello World" });
});

app.listen(8080, () => {
	console.log(`Server listening at http://localhost:8080`);
});
