const passwordInput = document.getElementById("password");
const showPassword = document.getElementById("showPassword");

if (passwordInput && showPassword) {
  showPassword.addEventListener("change", () => {
    passwordInput.type = showPassword.checked ? "text" : "password";
  });
}
