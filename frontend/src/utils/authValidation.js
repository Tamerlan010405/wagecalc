const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

function getPasswordByteLength(password){
  return new TextEncoder().encode(password).length;
}

export function validateEmail(value){
  const email = String(value ?? "").trim();

  if (email === ""){
    return "Введите e-mail";
  }

  if (!EMAIL_PATTERN.test(email)){
    return "Введите корректный e-mail";
  }

  return null;
}

export function validatePassword(value){
  const password = String(value ?? "");

  if (password.trim() === ""){
    return "Введите пароль";
  }

  if (password.length < 8){
    return "Пароль должен быть не короче 8 символов";
  }

  if (getPasswordByteLength(password) > 72){
    return "Пароль должен быть не длиннее 72 байт";
  }

  return null;
}

export function validateResetCode(value){
  const code = String(value ?? "").trim();

  if (code === ""){
    return "Введите код";
  }

  if (!/^\d{6}$/.test(code)){
    return "Код должен состоять из 6 цифр";
  }

  return null;
}
