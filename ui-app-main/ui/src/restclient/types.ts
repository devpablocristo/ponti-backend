export interface SuccessResponse<T> {
  success: true;
  message: string;
  data: T;
}

export interface ErrorResponse {
  error: {
    status: number;
    details: string;
  };
}

export class RequestError extends Error {
  private status: number | undefined;

  constructor(status: number | undefined, msg: string) {
    super(msg);
    this.status = status;
    Object.setPrototypeOf(this, RequestError.prototype);
  }

  getStatus(): number | undefined {
    return this.status;
  }
}
