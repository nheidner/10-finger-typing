import { Switch as HeadlessUiSwitch } from "@headlessui/react";
import classNames from "classnames";
import { FC } from "react";

export const Switch: FC<{
  item: string;
  label: string;
  enabled: boolean;
  handleChange: () => void;
}> = ({ item, label, enabled, handleChange }) => {
  return (
    <div className="flex items-center flex-col">
      <label
        htmlFor={item}
        className="block text-sm font-medium leading-6 text-gray-900 mb-2"
      >
        {label}
      </label>
      <HeadlessUiSwitch
        checked={enabled}
        onChange={handleChange}
        className={classNames(
          enabled ? "bg-indigo-600" : "bg-gray-200",
          "relative inline-flex h-9 w-14 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-indigo-600 focus:ring-offset-2"
        )}
      >
        <span className="sr-only">Use setting</span>
        <span
          aria-hidden="true"
          className={classNames(
            enabled ? "translate-x-5" : "translate-x-0",
            "pointer-events-none inline-block h-8 w-8 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out"
          )}
        />
      </HeadlessUiSwitch>
    </div>
  );
};
