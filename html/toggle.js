// Get references to the toggle button and the password input
const togglePassword = document.getElementById("togglePassword");
const passwordInput = document.getElementById("password");

togglePassword.addEventListener("click", () => {
  // Check if the password is currently hidden
  const isHidden = passwordInput.type === "password";

  // Switch between text and password
  passwordInput.type = isHidden ? "text" : "password";

  // Update the button text for sighted users
  togglePassword.textContent = isHidden ? "Hide" : "Show";

  // Update ARIA attributes for screen reader users
  togglePassword.setAttribute("aria-pressed", String(isHidden));
  togglePassword.setAttribute(
    "aria-label",
    isHidden ? "Hide password" : "Show password"
  );
});
