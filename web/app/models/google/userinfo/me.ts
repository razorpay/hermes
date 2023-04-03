import Model, { attr, hasMany } from "@ember-data/model";

export default class MeModel extends Model {
  @attr() declare email: string;
  @attr() declare given_name: string;
  @attr() declare name: string;
  @attr() declare picture: string;
}
