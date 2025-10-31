import dotenv from "dotenv";

dotenv.config();

class ConfigService {
  public readonly baseLoginApi = process.env.BASE_LOGIN_API as string;
  public readonly baseManagerApi = process.env.BASE_MANAGER_API as string;
  public readonly apiKey = process.env.X_API_KEY as string;
}

export const configService = new ConfigService();
