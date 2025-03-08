import { init } from "@paralleldrive/cuid2";
import { randomBytes } from "node:crypto";

export const instanceId = randomBytes(20).toString("hex");

export const generateId = init({
	length: 15,
	fingerprint: instanceId,
});
