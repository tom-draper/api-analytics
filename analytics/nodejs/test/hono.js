import { Hono } from 'hono';
import { honoAnalytics } from "../analytics.js";
import { serve } from '@hono/node-server';
import * as dotenv from "dotenv";
dotenv.config();

const apiKey = process.env.API_KEY;

const app = new Hono();

app.use('*', honoAnalytics(apiKey));

app.get('/', (c) => c.text('Hello, world!'));

serve(app, (info) => {
	console.log(`Server running at http://localhost:${info.port}`);
});