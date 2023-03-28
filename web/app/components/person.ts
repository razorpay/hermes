import Component from "@glimmer/component";

interface PersonComponentSignature {
  name?: string;
  email?: string;
  imgURL?: string;
  ignoreUnknown?: boolean;
}

export default class PersonComponent extends Component<PersonComponentSignature> {
  get isHidden() {
    return this.args.ignoreUnknown && !this.args.email;
  }
}
