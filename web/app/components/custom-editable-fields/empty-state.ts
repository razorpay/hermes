import Component from "@glimmer/component";

interface CustomEditableFieldsEmptyStateSignature {
  Args: {
    type?: "people" | "string" | "doc";
  };
}

export default class CustomEditableFieldsEmptyState extends Component<CustomEditableFieldsEmptyStateSignature> {
  get iconName(): string {
    switch (this.args.type) {
      case "people":
        return "user";
      case "doc":
        return "file-text";
      default:
        return "type";
    }
  }
}
