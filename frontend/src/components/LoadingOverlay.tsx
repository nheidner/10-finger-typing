import classNames from "classnames";
import { FC, ReactNode } from "react";

export const LoadingOverlay: FC<{
  isLoading: boolean;
  children: ReactNode;
}> = ({ isLoading, children }) => {
  return (
    <div className="relative">
      <div
        className={classNames(
          "absolute left-0 top-0 w-full h-full bg-white transition-opacity duration-100",
          isLoading ? "opacity-50" : "opacity-0 -z-10"
        )}
      />
      {children}
    </div>
  );
};
