export function validateFQDN(domain: string): {
  isValid: boolean;
  error?: string;
} {
  if (!domain || domain.trim() === "") {
    return { isValid: false, error: "Domain is required" };
  }

  const trimmedDomain = domain.trim();
  if (!trimmedDomain.endsWith(".")) {
    return {
      isValid: false,
      error:
        "Domain must end with a dot (.) to be a fully qualified domain name"
    };
  }

  const domainWithoutDot = trimmedDomain.slice(0, -1);
  if (domainWithoutDot.length === 0) {
    return { isValid: false, error: "Domain cannot be just a dot" };
  }

  if (domainWithoutDot.length > 253) {
    return {
      isValid: false,
      error: "Domain name is too long (max 253 characters)"
    };
  }

  const isWildcard = domainWithoutDot.startsWith("*.");
  const domainToValidate = isWildcard
    ? domainWithoutDot.slice(2)
    : domainWithoutDot;

  if (isWildcard && domainToValidate.length === 0) {
    return {
      isValid: false,
      error:
        "Wildcard domain must have at least one label after *. (e.g., *.example.com.)"
    };
  }

  // Check for valid characters and structure
  const domainRegex =
    /^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*$/;

  if (!domainRegex.test(domainToValidate)) {
    return {
      isValid: false,
      error:
        "Invalid domain format. Use only letters, numbers, dots, and hyphens"
    };
  }

  const labels = domainToValidate.split(".");
  for (const label of labels) {
    if (label.length > 63) {
      return {
        isValid: false,
        error: "Each part of the domain must be 63 characters or less"
      };
    }
    if (label.startsWith("-") || label.endsWith("-")) {
      return {
        isValid: false,
        error: "Domain parts cannot start or end with hyphens"
      };
    }
  }

  if (!domainToValidate.includes(".")) {
    return {
      isValid: false,
      error: "Domain must contain at least one dot (e.g., example.com.)"
    };
  }

  const tld = labels[labels.length - 1];
  if (!/^[a-zA-Z]{2,}$/.test(tld)) {
    return {
      isValid: false,
      error:
        "Top-level domain must contain only letters and be at least 2 characters"
    };
  }

  return { isValid: true };
}
