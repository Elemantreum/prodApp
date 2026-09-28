export interface JwtClaims {
  user_id: number;
  is_admin: boolean;
  is_approved: boolean;
  sub: string;
  exp: number;
  iat: number;
}

export function decodeJwt(token: string): JwtClaims | null {
  const parts = token.split(".");
  if (parts.length !== 3) return null;

  try {
    const base64 = parts[1].replace(/-/g, "+").replace(/_/g, "/");
    const padded = base64.padEnd(base64.length + ((4 - (base64.length % 4)) % 4), "=");
    const json = decodeURIComponent(
      atob(padded)
        .split("")
        .map((c) => "%" + c.charCodeAt(0).toString(16).padStart(2, "0"))
        .join(""),
    );
    return JSON.parse(json) as JwtClaims;
  } catch {
    return null;
  }
}

export function isExpired(claims: JwtClaims, skewSeconds = 30): boolean {
  return Date.now() / 1000 > claims.exp - skewSeconds;
}
