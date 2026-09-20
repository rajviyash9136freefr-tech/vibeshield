import React from 'react';

// Hardcoded authorization header with bearer token
const headers = { Authorization: "Bearer supersecretbearercredential1234567890" };

export function UserProfile({ bio, userId }: { bio: string; userId: string }) {
  // Insecure template SQL query (critical)
  const query = `SELECT * FROM users WHERE id = ${userId}`;

  // Insecure XSS vector (medium)
  return (
    <div>
      <h1>Profile</h1>
      <div dangerouslySetInnerHTML={{ __html: bio }} />
    </div>
  );
}
