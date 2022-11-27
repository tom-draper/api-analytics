import express from 'express';
import analytics from '../analytics.js';
import * as dotenv from 'dotenv';
dotenv.config();

const apiKey = process.env.API_KEY

const app = express()

app.use(analytics(apiKey))

app.get("/", (req, res) => {
    res.send({message: "Hello World"});
});

app.listen(5000, () => {
    console.log(`Running on PORT 5000`);
})