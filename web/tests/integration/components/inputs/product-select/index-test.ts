import { module, test } from "qunit";
import { setupRenderingTest } from "ember-qunit";
import { hbs } from "ember-cli-htmlbars";
import { click, render, waitFor } from "@ember/test-helpers";
import { setupMirage } from "ember-cli-mirage/test-support";
import { MirageTestContext } from "ember-cli-mirage/test-support";
import { BadgeSize } from "hermes/types/hds-badge";
import { Placement } from "@floating-ui/dom";

const DEFAULT_DROPDOWN_SELECTOR =
  "[data-test-product-select-default-dropdown-toggle]";

const LIST_ITEM_SELECTOR = "[data-test-product-select-item]";

interface InputsProductSelectContext extends MirageTestContext {
  selected?: any;
  onChange: (value: string) => void;
  badgeSize?: BadgeSize;
  formatIsBadge?: boolean;
  placement?: Placement;
  isSaving?: boolean;
}

module("Integration | Component | inputs/product-select", function (hooks) {
  setupRenderingTest(hooks);
  setupMirage(hooks);

  hooks.beforeEach(async function (this: InputsProductSelectContext) {
    this.server.createList("product", 3);
    this.set("selected", "Vault");
    this.set("onChange", () => {});
  });

  test("it can render in two formats", async function (this: InputsProductSelectContext, assert) {
    const badgeDropdownSelector = "[data-test-badge-dropdown-list]";

    this.set("formatIsBadge", true);

    await render(hbs`
      {{! @glint-nocheck: not typesafe yet }}
      <Inputs::ProductSelect
        @selected={{this.selected}}
        @onChange={{this.onChange}}
        @formatIsBadge={{this.formatIsBadge}}
      />
    `);

    assert.dom(badgeDropdownSelector).exists("badge dropdown is rendered");
    assert
      .dom(DEFAULT_DROPDOWN_SELECTOR)
      .doesNotExist("default dropdown is not rendered");

    this.set("formatIsBadge", false);

    assert
      .dom(badgeDropdownSelector)
      .doesNotExist("badge dropdown is not rendered");
    assert
      .dom(DEFAULT_DROPDOWN_SELECTOR)
      .exists("default dropdown is rendered");
  });

  test("it can render the toggle with a product abbreviation", async function (this: InputsProductSelectContext, assert) {
    this.set("selected", this.server.schema.products.first().name);

    await render(hbs`
      {{! @glint-nocheck: not typesafe yet }}
      <Inputs::ProductSelect
        @selected={{this.selected}}
        @onChange={{this.onChange}}
      />
    `);

    assert
      .dom("[data-test-product-select-toggle-abbreviation]")
      .hasText("TST-0");
  });

  test("it shows an empty state when nothing is selected (default toggle)", async function (this: InputsProductSelectContext, assert) {
    this.set("selected", undefined);

    await render(hbs`
      {{! @glint-nocheck: not typesafe yet }}
      <Inputs::ProductSelect
        @selected={{this.selected}}
        @onChange={{this.onChange}}
      />
    `);
    assert.dom("[data-test-product-select-selected-value]").hasText("--");
  });

  test("it displays the products in a dropdown list with abbreviations", async function (this: InputsProductSelectContext, assert) {
    await render(hbs`
      {{! @glint-nocheck: not typesafe yet }}
      <Inputs::ProductSelect
        @selected={{this.selected}}
        @onChange={{this.onChange}}
      />
    `);

    await click(DEFAULT_DROPDOWN_SELECTOR);

    assert.dom(LIST_ITEM_SELECTOR).exists({ count: 3 });

    let firstListItem = this.element.querySelector(LIST_ITEM_SELECTOR);
    assert.dom(firstListItem).hasText("Test Product 0 TST-0");
  });

  test("it fetches the products if they aren't already loaded", async function (this: InputsProductSelectContext, assert) {
    this.server.db.emptyData();

    await render(hbs`
      {{! @glint-nocheck: not typesafe yet }}
      <Inputs::ProductSelect
        @selected={{this.selected}}
        @onChange={{this.onChange}}
      />
    `);

    await click(DEFAULT_DROPDOWN_SELECTOR);

    // In Mirage, when there are no products, we return a single default product
    assert.dom(LIST_ITEM_SELECTOR).exists({ count: 1 });
    assert.dom(LIST_ITEM_SELECTOR).hasText("Default Fetched Product NONE");
  });
});
