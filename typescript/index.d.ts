/* eslint-disable */
/**
 * This file was automatically generated by json-schema-to-typescript.
 * DO NOT MODIFY IT BY HAND. Instead, modify the source JSONSchema file,
 * and run json-schema-to-typescript to regenerate this file.
 */

export type PageSchema = List | Detail;
export type Action =
  | {
      /**
       * The type of the action.
       */
      type: "copy-text";
      /**
       * The title of the action.
       */
      title?: string;
      /**
       * The text to copy.
       */
      text: string;
      /**
       * The key used as a shortcut.
       */
      key?: string;
    }
  | {
      /**
       * The type of the action.
       */
      type: "open-file";
      /**
       * The title of the action.
       */
      title?: string;
      /**
       * The key used as a shortcut.
       */
      key?: string;
      /**
       * The path to open.
       */
      path?: string;
    }
  | {
      /**
       * The type of the action.
       */
      type: "open-url";
      /**
       * The title of the action.
       */
      title?: string;
      /**
       * The key used as a shortcut.
       */
      key?: string;
      /**
       * The url to open.
       */
      url?: string;
    }
  | {
      /**
       * The type of the action.
       */
      type: "run-command";
      /**
       * The title of the action.
       */
      title?: string;
      /**
       * The key used as a shortcut.
       */
      key?: string;
      /**
       * The command to run.
       */
      command: string;
      /**
       * The input to pass to the command stdin.
       */
      input?: string;
      /**
       * The directory where the command should be run.
       */
      dir?: string;
      /**
       * The action to take when the command succeeds.
       */
      onSuccess?: "reload" | "exit" | "push";
      /**
       * The inputs to show when the action is run.
       */
      inputs?: Input[];
    }
  | {
      /**
       * The type of the action.
       */
      type: "read-file";
      /**
       * The title of the action.
       */
      title?: string;
      /**
       * The key used as a shortcut.
       */
      key?: string;
      /**
       * The path to read.
       */
      path: string;
    }
  | {
      type: "push-page";
      page:
        | string
        | {
            [k: string]: unknown;
          };
    };
export type Input =
  | {
      /**
       * The name of the input.
       */
      name: string;
      /**
       * The title of the input.
       */
      title: string;
      /**
       * The type of the input.
       */
      type: "textfield";
      /**
       * The placeholder of the input.
       */
      placeholder?: string;
      /**
       * The default value of the input.
       */
      default?: string;
      /**
       * Whether the input should be secure.
       */
      secure?: boolean;
    }
  | {
      /**
       * The name of the input.
       */
      name: string;
      /**
       * The title of the input.
       */
      title: string;
      /**
       * The type of the input.
       */
      type: "checkbox";
      /**
       * The default value of the input.
       */
      default?: boolean;
      /**
       * The label of the input.
       */
      label?: string;
      /**
       * The text substitution to use when the input is true.
       */
      trueSubstitution?: string;
      /**
       * The text substitution to use when the input is false.
       */
      falseSubstitution?: string;
    }
  | {
      /**
       * The name of the input.
       */
      name: string;
      /**
       * The title of the input.
       */
      title: string;
      /**
       * The type of the input.
       */
      type: "textarea";
      /**
       * The placeholder of the input.
       */
      placeholder?: string;
      /**
       * The default value of the input.
       */
      default?: string;
    }
  | {
      /**
       * The name of the input.
       */
      name: string;
      /**
       * The title of the input.
       */
      title: string;
      /**
       * The type of the input.
       */
      type: "dropdown";
      /**
       * The items of the input.
       */
      items: {
        /**
         * The title of the item.
         */
        title: string;
        /**
         * The value of the item.
         */
        value: string;
      }[];
      /**
       * The default value of the input.
       */
      default?: string;
    };
/**
 * The preview to show in the detail view.
 */
export type Preview =
  | {
      /**
       * The text to show in the preview.
       */
      text: string;
      /**
       * The language of the preview text.
       */
      language?: string;
    }
  | {
      /**
       * The command used to generate the preview.
       */
      command: string;
      /**
       * The directory where the command should be run.
       */
      dir?: string;
      /**
       * The language of the preview text.
       */
      language?: string;
    };

export interface List {
  /**
   * The type of the response.
   */
  type: "list";
  /**
   * The title of the page.
   */
  title?: string;
  /**
   * The text to show when the list is empty.
   */
  emptyText?: string;
  /**
   * Whether to show details on the right side of the list.
   */
  showPreview?: boolean;
  /**
   * The global actions attached to the list.
   */
  actions?: Action[];
  /**
   * The items in the list.
   */
  items?: Listitem[];
}
export interface Listitem {
  /**
   * The title of the item.
   */
  title: string;
  /**
   * The id of the item.
   */
  id?: string;
  /**
   * The subtitle of the item.
   */
  subtitle?: string;
  preview?: Preview;
  /**
   * The accessories to show on the right side of the item.
   */
  accessories?: string[];
  /**
   * The actions attached to the item.
   */
  actions?: Action[];
}
/**
 * A detail view displayign a preview and actions.
 */
export interface Detail {
  /**
   * The type of the response.
   */
  type: "detail";
  /**
   * The title of the page.
   */
  title?: string;
  preview?: Preview;
  /**
   * The actions attached to the detail view.
   */
  actions?: Action[];
}
