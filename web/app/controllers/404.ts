import Controller from "@ember/controller";
import { HdsIconPosition } from "hermes/enums/hds-components";
import parseDate from "hermes/utils/parse-date";

export default class Error404Controller extends Controller {
  HdsIconPosition = HdsIconPosition;

  get currentDate() {
    return parseDate(new Date(), "long");
  }
}
