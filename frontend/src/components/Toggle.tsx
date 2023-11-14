import { ChangeEvent, FC } from "react";

export const Toggle: FC<{
  item: string;
  label: string;
  options: string[];
  selectedValue: string;
  handleChange: (e: ChangeEvent<HTMLSelectElement>) => void;
}> = ({ item, label, options, selectedValue, handleChange }) => {
  return (
    <div className="flex items-center flex-col">
      <label
        htmlFor={item}
        className="block text-sm font-medium leading-6 text-gray-900"
      >
        {label}
      </label>
      <select
        id={item}
        name={item}
        className="mt-2 block w-full rounded-md border-0 py-1.5 pl-3 pr-10 text-gray-900 ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-indigo-600 sm:text-sm sm:leading-6"
        onChange={handleChange}
        value={selectedValue}
      >
        {options.map((option, key) => {
          return (
            <option key={key} value={option}>
              {option}
            </option>
          );
        })}
      </select>
    </div>
  );
};
