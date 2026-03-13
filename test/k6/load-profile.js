import http from "k6/http";
import {check, sleep} from "k6";

export const options = {
    vus: 1000,
    duration: "30s",
};

const actions = ["PUT", "GET"]
const randomInt = (max) => {
    return Math.floor(Math.random() * max);
}

const randomKey = (length = 5) => {
    const chars = "abcdefghijklmnopqrstuvwxyz";
    let result = "";

    for (let i = 0; i < length; i++) {
        const idx = Math.floor(Math.random() * chars.length);
        result += chars[idx];
    }

    return result;
}

export default function () {
    const randomValue = actions[Math.floor(Math.random() * actions.length)];
    let res = null
    if (randomValue === "PUT") {
        const key = randomKey();
        const value = randomInt(10000);
        const url = `http://localhost:3000/put/${key}`
        res = http.post(url, `${value}`)

    } else {
        const key = randomKey();
        const url = `http://localhost:3000/get/${key}`
        res = http.get(url);
    }
    check(res, {"status is 200": (r) => {
        return r.status === 200
        },});
}