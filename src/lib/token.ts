import { jwtDecode } from 'jwt-decode';

type JwtClaims = {
	userId: string;
	exp: number;
};

type JwtRefreshClaims = {
	userId: string;
	exp: number;
};

function isExpired(token: string): boolean {
	let decoded: JwtClaims | JwtRefreshClaims;
	try {
		decoded = jwtDecode(token) as JwtClaims | JwtRefreshClaims;
	} catch (err) {
		return true;
	}
	return Date.now() >= decoded.exp * 1000;
}

export { isExpired };
export type { JwtClaims, JwtRefreshClaims };
