import winston from "winston";
import { env } from "./env.js";

const colors = {
	error: "red",
	warn: "yellow",
	info: "green",
	http: "magenta",
	debug: "blue",
};

const jsonFormat = winston.format.combine(
	winston.format.timestamp({ format: "YYYY-MM-DD HH:mm:ss:ms" }),
	winston.format.json(),
	winston.format.errors({ stack: true }),
);

const devConsoleFormat = winston.format.combine(
	winston.format.colorize({ all: true }),
	winston.format.timestamp({ format: "YYYY-MM-DD HH:mm:ss:ms" }),
	winston.format.printf(
		(info) =>
			`${info.timestamp} ${info.level}: ${info.message} ${info.stack || ""} ${Object.keys(info.metadata || {}).length ? JSON.stringify(info.metadata || {}) : ""}`,
	),
);

winston.addColors(colors);

export const logger = winston.createLogger({
	level: "info",
	format: env.NODE_ENV === "development" ? devConsoleFormat : jsonFormat,
	transports: [new winston.transports.Console()],
});
