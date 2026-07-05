import Keycloak from "keycloak-js";

export const keycloak = new Keycloak({
  url: import.meta.env.VITE_KEYCLOAK_URL ?? "http://localhost:8080",
  realm: import.meta.env.VITE_KEYCLOAK_REALM ?? "onboarder",
  clientId: import.meta.env.VITE_KEYCLOAK_CLIENT_ID ?? "keycloak-onboarder-ui",
});

export async function initKeycloak() {
  const authenticated = await keycloak.init({
    onLoad: "login-required",
    pkceMethod: "S256",
    checkLoginIframe: false,
  });

  return authenticated;
}

export async function refreshToken() {
  if (!keycloak.token) return;

  await keycloak.updateToken(30);
}

export function getAccessToken() {
  return keycloak.token;
}

export function logout() {
  return keycloak.logout({
    redirectUri: window.location.origin,
  });
}
