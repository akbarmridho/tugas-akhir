import crypto from "node:crypto";

export const computeHMACSHA256 = (secret: string, payload: string) => {
	const hmac = crypto.createHmac("sha256", secret);

	hmac.update(payload);

	return hmac.digest("hex");
};
