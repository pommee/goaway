import {
  RuleConfigCondition,
  RuleConfigSeverity,
  TargetCaseType,
  UserConfig
} from "@commitlint/types";

const config: UserConfig = {
  parserPreset: "conventional-changelog-conventionalcommits",
  rules: {
    "body-leading-blank": [RuleConfigSeverity.Warning, "always"] as const,
    "body-max-line-length": [
      RuleConfigSeverity.Error,
      "always",
      999999
    ] as const,
    "footer-leading-blank": [RuleConfigSeverity.Warning, "always"] as const,
    "footer-max-line-length": [
      RuleConfigSeverity.Error,
      "always",
      100
    ] as const,
    "header-max-length": [RuleConfigSeverity.Error, "always", 100] as const,
    "header-trim": [RuleConfigSeverity.Error, "always"] as const,
    "subject-case": [
      RuleConfigSeverity.Error,
      "never",
      ["sentence-case", "start-case", "pascal-case", "upper-case"]
    ] as [RuleConfigSeverity, RuleConfigCondition, TargetCaseType[]],
    "subject-empty": [RuleConfigSeverity.Error, "never"] as const,
    "type-case": [RuleConfigSeverity.Error, "always", "lower-case"] as const,
    "type-empty": [RuleConfigSeverity.Error, "never"] as const,
    "type-enum": [
      RuleConfigSeverity.Error,
      "always",
      [
        "build",
        "chore",
        "ci",
        "docs",
        "feat",
        "fix",
        "perf",
        "refactor",
        "revert",
        "style",
        "ui",
        "test"
      ]
    ] as [RuleConfigSeverity, RuleConfigCondition, string[]]
  }
};

export default config;
