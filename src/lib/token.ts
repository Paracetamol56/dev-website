import jwt from "jsonwebtoken";

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
    decoded = jwt.decode(token) as JwtClaims | JwtRefreshClaims;
  } catch (err) {
    return true;
  }
  return Date.now() >= decoded.exp * 1000;
}

export { isExpired };
export type { JwtClaims, JwtRefreshClaims };

