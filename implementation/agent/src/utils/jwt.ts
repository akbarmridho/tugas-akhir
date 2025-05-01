import { v4 as uuidv4 } from "https://jslib.k6.io/uuid/1.0.0/index.js";
import encoding from "k6/encoding";
import crypto from "k6/crypto";

function base64urlEncode(data: string | ArrayBuffer) {
	return encoding.b64encode(data, "rawurl");
}

interface JWTPayload {
	userId: string;
	jwtToken: string;
}

export function forgeJwt(): JWTPayload {
	const header = {
		alg: "HS256",
		typ: "JWT",
	};

	const userId = uuidv4();

	const payload = {
		user_id: userId,
		iss: "k6",
		sub: "k6-user",
		aud: ["ticket-service"],
		exp: Math.floor(Date.now() / 1000) + 10800, // expires in 3h
		nbf: Math.floor(Date.now() / 1000),
		iat: Math.floor(Date.now() / 1000),
		jti: uuidv4(),
	};

	const encodedHeader = base64urlEncode(JSON.stringify(header));
	const encodedPayload = base64urlEncode(JSON.stringify(payload));

	const dataToSign = `${encodedHeader}.${encodedPayload}`;
	const signature = crypto.hmac("sha256", "secret", dataToSign, "hex");
	const encodedSignature = base64urlEncode(hexToBytes(signature).buffer);

	return {
		userId: userId,
		jwtToken: `${dataToSign}.${encodedSignature}`,
	};
}

// Helper to convert hex to bytes
function hexToBytes(hex: string) {
	const bytes = [];
	for (let c = 0; c < hex.length; c += 2) {
		bytes.push(parseInt(hex.substring(c, 2), 16));
	}
	return new Uint8Array(bytes);
}
