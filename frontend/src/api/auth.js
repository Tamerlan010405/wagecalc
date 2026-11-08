const API_BASE_URL = import.meta.env.VITE_API_URL ?? "";

export class ApiError extends Error{
  constructor(message, details = {}){
    super(message);
    this.name = "ApiError";
    this.status = details.status;
    this.code = details.code;
    this.field = details.field;
  }
}

async function getResponseBody(response){
  try{
    return await response.json();
  }catch{
    return {};
  }
}

async function authRequest(path, options = {}){
  const requestOptions = {
    method: options.method ?? "GET",
    credentials: "include",
  };

  if (options.body !== undefined){
    requestOptions.headers = {
      "Content-Type": "application/json",
    };
    requestOptions.body = JSON.stringify(options.body);
  }

  const response = await fetch(`${API_BASE_URL}${path}`, requestOptions);
  const responseBody = await getResponseBody(response);

  if (!response.ok){
    const error = responseBody?.error ?? {};
    throw new ApiError(error.message ?? `Ошибка запроса: ${response.status}`, {
      status: response.status,
      code: error.code,
      field: error.field,
    });
  }

  return responseBody;
}

export function registerUser({ email, password }){
  return authRequest("/api/v1/auth/register", {
    method: "POST",
    body: {
      email,
      password,
    },
  });
}

export function loginUser({ email, password }){
  return authRequest("/api/v1/auth/login", {
    method: "POST",
    body: {
      email,
      password,
    },
  });
}

export function logoutUser(){
  return authRequest("/api/v1/auth/logout", {
    method: "POST",
  });
}

export function requestPasswordReset({ email }){
  return authRequest("/api/v1/auth/password-reset/request", {
    method: "POST",
    body: {
      email,
    },
  });
}

export function confirmPasswordReset({ email, code, newPassword }){
  return authRequest("/api/v1/auth/password-reset/confirm", {
    method: "POST",
    body: {
      email,
      code,
      new_password: newPassword,
    },
  });
}
