import { http, HttpResponse } from "msw";
import jwt from "jsonwebtoken";

import { configService } from "../configService";

interface LoginRequest {
  username: string;
  password: string;
}

const ACCESS_TOKEN_EXPIRATION = 15 * 60;
const REFRESH_TOKEN_EXPIRATION = 6 * 30 * 24 * 60 * 60;

const MOCK_USER = {
  id: 1,
  rolId: 1,
  username: "testuser",
  password: "123456",
  email: "testuser@example.com",
  tokenHash: "randomhash123",
};

export const handlers = [
  http.post(configService.baseLoginApi + "/auth/login", async ({ request }) => {
    try {
      const body = (await request.json()) as LoginRequest;
      if (!body.username || !body.password) {
        return new HttpResponse(
          JSON.stringify({ message: "Usuario y contraseña requeridos" }),
          {
            status: 400,
            headers: {
              "Content-Type": "application/json",
            },
          }
        );
      }

      if (
        body.username !== MOCK_USER.username ||
        body.password !== MOCK_USER.password
      ) {
        return new HttpResponse(
          JSON.stringify({ message: "Credenciales inválidas" }),
          { status: 401, headers: { "Content-Type": "application/json" } }
        );
      }

      const now = Math.floor(Date.now() / 1000);
      const accessExp = now + ACCESS_TOKEN_EXPIRATION;
      const refreshExp = now + REFRESH_TOKEN_EXPIRATION;

      const accessToken = generateToken({
        id: MOCK_USER.id,
        rolId: MOCK_USER.rolId,
        username: MOCK_USER.username,
        KeyType: "access",
        exp: accessExp,
        iss: "auth.service",
      });

      const refreshToken = generateToken({
        id: MOCK_USER.id,
        hash: MOCK_USER.tokenHash,
        keyType: "refresh",
        exp: refreshExp,
        iss: "auth.service",
      });

      return new Response(
        JSON.stringify({
          access_token: accessToken,
          refresh_token: refreshToken,
        }),
        {
          headers: {
            "Content-Type": "application/json",
          },
        }
      );
    } catch (error) {
      return new Response(
        JSON.stringify({ message: "Error interno del servidor" }),
        { status: 500, headers: { "Content-Type": "application/json" } }
      );
    }
  }),
  http.get(configService.baseLoginApi + "/auth/protected/hi", ({ request }) => {
    const authHeader = request.headers.get("authorization");
    if (!authHeader) {
      return new HttpResponse(JSON.stringify({ message: "Invalid token" }), {
        status: 401,
        headers: {
          "Content-Type": "application/json",
        },
      });
    }

    const token = authHeader.split(" ")[1];
    if (!token) {
      return new HttpResponse(JSON.stringify({ message: "Invalid token" }), {
        status: 401,
        headers: {
          "Content-Type": "application/json",
        },
      });
    }

    const secretKey = process.env.JWT_SECRET;
    if (!secretKey) {
      return new HttpResponse(
        JSON.stringify({ message: "Internal Server Error" }),
        {
          status: 500,
          headers: {
            "Content-Type": "application/json",
          },
        }
      );
    }

    try {
      jwt.verify(token, secretKey);
      return new Response(
        JSON.stringify({
          message: "Login successful",
          access_token: token,
        }),
        {
          headers: {
            "Content-Type": "application/json",
          },
        }
      );
    } catch (err) {
      return new HttpResponse(JSON.stringify({ message: "Invalid token" }), {
        status: 401,
        headers: {
          "Content-Type": "application/json",
        },
      });
    }
  }),
];

const generateToken = (claims: object): string => {
  const secretKey = process.env.JWT_SECRET;
  if (!secretKey) {
    return "";
  }

  return jwt.sign(claims, secretKey, { algorithm: "HS256" });
};
