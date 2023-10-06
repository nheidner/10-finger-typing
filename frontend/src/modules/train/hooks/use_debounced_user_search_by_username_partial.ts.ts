import { debounce } from "@/utils/debounce";
import { getUsersByUsernamePartial } from "@/utils/queries";
import { useQuery } from "@tanstack/react-query";
import { ChangeEvent, useRef, useState } from "react";

export const useDebouncedUserSearchByUsernamePartial = () => {
  const [queryKey, setQueryKey] = useState("");
  const debouncedSetQueryKeyRef = useRef(debounce<void>(setQueryKey, 300));

  const handleUsernamePartialChange = (
    event: ChangeEvent<HTMLInputElement>
  ) => {
    const { value } = event.target;

    debouncedSetQueryKeyRef.current(value);
  };

  const { data } = useQuery({
    queryKey: ["users", "username_contains", queryKey],
    queryFn: () => getUsersByUsernamePartial(queryKey),
    retry: false,
    enabled: !!queryKey,
  });

  return {
    users: data,
    handleUsernamePartialChange,
  };
};
