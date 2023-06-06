// https://helios.hashicorp.design/components/button?tab=code#component-api

import { ComponentLike } from "@glint/template";
import {
  HdsButtonColor,
  HdsButtonIconPosition,
  HdsButtonSize,
} from "hermes/types/hds-components";

export type HdsButtonComponent = ComponentLike<{
  Element: HTMLButtonElement;
  Args: {
    text: string;
    size?: HdsButtonSize;
    color?: HdsButtonColor;
    icon?: string;
    iconPosition?: HdsButtonIconPosition;
    isIconOnly?: boolean;
    isFullWidth?: boolean;
    href?: string;
    isHrefExternal?: boolean;
    isRouteExternal?: boolean;
    route?: string;
    models?: unknown[];
    model?: unknown;
    query?: Record<string, unknown>;
    "current-when"?: string;
    replace?: boolean;
  };
}>;
