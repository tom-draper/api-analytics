import Koa from "koa";
import { koaAnalytics } from "../analytics.js";
import * as dotenv from "dotenv";
dotenv.config();

const apiKey = process.env.API_KEY;

const app = new Koa();

app.use(koaAnalytics(apiKey));

app.use((ctx) => {
	ctx.body = { message: "Hello World!" };
});

app.listen(8080, () =>
	console.log("Server listening at http://localhost:8080")
);
