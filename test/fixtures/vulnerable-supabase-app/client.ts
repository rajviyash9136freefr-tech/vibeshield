import { createClient } from '@supabase/supabase-js';
import jwt from 'jsonwebtoken';

const SUPABASE_URL = "https://myproject.supabase.co";
// Hardcoded service role key bypasses Row Level Security completely
const api_key = "FAKE_SUPABASE_SERVICE_ROLE_KEY_FOR_TESTING_PURPOSES_ONLY";

export const supabase = createClient(SUPABASE_URL, api_key);

export function decodeToken(token: string) {
  // Insecure JWT decode with 'none' algorithm
  return jwt.verify(token, "secret", { algorithms: ["none", "HS256"] });
}
