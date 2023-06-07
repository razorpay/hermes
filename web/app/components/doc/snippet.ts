import Component from "@glimmer/component";

interface DocSnippetComponentSignature {
  Element: HTMLParagraphElement;
  Args: {
    snippet: string;
  };
}

export default class DocSnippetComponent extends Component<DocSnippetComponentSignature> {}
