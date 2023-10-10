import { debounce } from "@/utils/debounce";
import { getUsersByUsernamePartial } from "@/utils/queries";
import { useQuery } from "@tanstack/react-query";
import { ChangeEvent, useRef, useState } from "react";

export const useDebouncedUserSearchByUsernamePartial = (
  debounceTime: number
) => {
  const [queryKey, setQueryKey] = useState("");
  const debouncedSetQueryKeyRef = useRef(
    debounce<void>(setQueryKey, debounceTime)
  );

  const handleUsernamePartialChange = (usernamePartial: string) => {
    debouncedSetQueryKeyRef.current(usernamePartial);
  };

  const { data } = useQuery({
    queryKey: ["users", "username_contains", queryKey],
    queryFn: () => getUsersByUsernamePartial(queryKey),
    retry: false,
    enabled: !!queryKey,
    keepPreviousData: true,
  });

  return {
    users: data,
    debouncedFetchUsers: handleUsernamePartialChange,
  };
};
