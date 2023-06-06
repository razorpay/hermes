import Component from "@glimmer/component";
import { HdsButtonColor, HdsIconPosition } from "hermes/enums/hds-components";

interface XDropdownListToggleButtonComponentSignature {
  Element: HTMLButtonElement;
  Args: {
    registerAnchor: () => void;
    onTriggerKeydown: () => void;
    toggleContent: () => void;
    contentIsShown: boolean;
    disabled?: boolean;
    ariaControls: string;
    icon?: string;
    color?: HdsButtonColor;
    iconPosition?: HdsIconPosition;
    text: string;
  };
  Blocks: {
    default: [];
  };
}

export default class XDropdownListToggleButtonComponent extends Component<XDropdownListToggleButtonComponentSignature> {
  HdsIconPosition = HdsIconPosition;
}

declare module "@glint/environment-ember-loose/registry" {
  export default interface Registry {
    "x/dropdown-list/toggle-button": typeof XDropdownListToggleButtonComponent;
  }
}
