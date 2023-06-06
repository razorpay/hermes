//helios.hashicorp.design/components/badge-count?tab=code#component-api

import { ComponentLike } from "@glint/template";
import {
  HdsBadgeCountColor,
  HdsBadgeCountType,
} from "hermes/enums/hds-components";
import { HdsBadgeCountSize } from "hermes/types/HdsBadgeCountSize";

export type HdsBadgeCountComponent = ComponentLike<{
  Element: HTMLDivElement;
  Args: {
    text: string;
    size?: HdsBadgeCountSize;
    type?: HdsBadgeCountType;
    color?: HdsBadgeCountColor;
  };
}>;
