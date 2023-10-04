import classNames from "classnames";
import { SortByOption } from "../types";
import { sortByOptions } from "../constants";

export const SortBy = ({
  sortBy,
  setSortBy,
}: {
  sortBy: SortByOption;
  setSortBy: (sortByOption: SortByOption) => void;
}) => {
  const selectOption = (sortByOption: SortByOption) => () =>
    setSortBy(sortByOption);

  return (
    <div className="flex items-center">
      <h4 className="text-sm font-semibold leading-7 text-gray-700">
        Sort by:
      </h4>
      {sortByOptions
        .map((sortByOption, index) => {
          const isLast = index === sortByOptions.length - 1;
          const isActive = sortByOption === sortBy;

          const option = (
            <span
              onClick={selectOption(sortByOption)}
              key={sortByOption}
              className={classNames(
                "inline-block mx-2 text-sm cursor-pointer",
                isActive && "font-semibold text-indigo-400"
              )}
            >
              {sortByOption}
            </span>
          );

          if (isLast) {
            return option;
          }

          return [
            option,
            <span key={index} className="text-gray-500">
              |
            </span>,
          ];
        })
        .flat()}
    </div>
  );
};
