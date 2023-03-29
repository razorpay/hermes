import Component from "@glimmer/component";

interface PersonComponentSignature {
  name?: string;
  email?: string;
  imgURL?: string;
  ignoreUnknown?: boolean;
  hideEmail?: boolean;
}

export default class PersonComponent extends Component<PersonComponentSignature> {
  get isHidden() {
    return this.args.ignoreUnknown && !this.args.email;
  }

  get emailIsUnknown() {
    return !this.args.email;
  }

  get primaryText() {
    return this.args.name || this.args.email || "Unknown";
  }

  get secondaryTextIsHidden() {
    return (
      this.args.hideEmail ||
      this.emailIsUnknown ||
      this.primaryText === this.args.email
    );
  }
}
