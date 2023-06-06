// https://helios.hashicorp.design/components/button?tab=code#component-api

import { ComponentLike } from "@glint/template";
import {
  HdsButtonColor,
  HdsIconPosition,
  HdsComponentSize,
} from "hermes/enums/hds-components";

export type HdsButtonComponent = ComponentLike<{
  Element: HTMLButtonElement;
  Args: {
    text: string;
    size?: HdsComponentSize;
    color?: HdsButtonColor;
    icon?: string;
    iconPosition?: HdsIconPosition;
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
