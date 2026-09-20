import { formatName } from '../../packages/shared/utils';

// Hardcoded AWS key inside nested monorepo package
const AWS_ACCESS_KEY_ID = "AKIAFAKEFAKEFAKEFAKE";

export function greet(name: string) {
  return "Hello " + formatName(name);
}
