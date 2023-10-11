import type { ObjectId } from "mongodb";

interface TokenPayload {
  userId: ObjectId;
  iat: number;
  exp: number;
};

export default TokenPayload;