import { module, test } from "qunit";
import cleanString from "hermes/utils/clean-string";

module("Unit | Utility | clean-string", function () {
  test("it correctly cleans strings", function (assert) {
    assert.equal(cleanString("\u2018\u2019\u201A"), "'''");
    assert.equal(cleanString("\u201C\u201D\u201E"), '"""');
    assert.equal(cleanString("\u2026"), "...");
    assert.equal(cleanString("\u2013"), "-");
    assert.equal(cleanString("\u2014"), "--");
    assert.equal(cleanString("\u2022\u00B7"), "**");
    assert.equal(cleanString("\u00AB"), "<<");
    assert.equal(cleanString("\u00BB"), ">>");
    assert.equal(cleanString("\u2039"), "<");
    assert.equal(cleanString("\u203A"), ">");
    assert.equal(cleanString("\u00A2"), "c");
    assert.equal(cleanString("\u00A9"), "(C)");
    assert.equal(cleanString("\u2122"), "(TM)");
    assert.equal(cleanString("\u00AE"), "(R)");
    assert.equal(cleanString("\u00B0"), "(deg)");
    assert.equal(cleanString("\u00B1"), "+/-");
    assert.equal(cleanString("\u00D7"), "*");
    assert.equal(cleanString("\u00F7"), "/");
    assert.equal(cleanString("\u00BC"), "1/4");
    assert.equal(cleanString("\u00BD"), "1/2");
    assert.equal(cleanString("\u00BE"), "3/4");
    assert.equal(cleanString("\u221A"), "sqrt");
    assert.equal(cleanString("\u00B9"), "^1");
    assert.equal(cleanString("\u00B2"), "^2");
    assert.equal(cleanString("\u00B3"), "^3");
    assert.equal(cleanString("\u207F"), "^n");
    assert.equal(cleanString("\u00BF"), "?");
    assert.equal(cleanString("\u00A1"), "!");
    assert.equal(cleanString("\u2265"), ">=");
    assert.equal(cleanString("\u2264"), "<=");
    assert.equal(cleanString("\u2260"), "!=");
    assert.equal(cleanString("pøkémöñ"), "pokemon");
    assert.equal(cleanString("∑π¬µ∫≈π†ºª¶§∞£"), "");
  });
});