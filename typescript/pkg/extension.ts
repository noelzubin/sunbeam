import type { Action } from "./action.ts";

export type Extension = {
  name: string;
} & Manifest;

export type Manifest = {
  title: string;
  description?: string;
  commands?: readonly Command[];
  root?: Action[];
};

export type Command = {
  name: string;
  description?: string;
  params?: readonly ParamDef[];
  mode: "filter" | "search" | "detail" | "silent";
};

export type ParamDef = {
  name: string;
  type: "string" | "number" | "boolean";
  description?: string;
  optional?: boolean;
};

