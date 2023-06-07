import Component from "@glimmer/component";
import { HdsButtonColor, HdsIconPosition } from "hermes/enums/hds-components";

interface XDropdownListToggleButtonComponentSignature {
  Element: HTMLButtonElement;
  Args: {
    registerAnchor: (element: HTMLElement) => void;
    onTriggerKeydown: (event: KeyboardEvent) => void;
    toggleContent: () => void;
    contentIsShown: boolean;
    disabled?: boolean;
    ariaControls: string;
    color: HdsButtonColor;
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
