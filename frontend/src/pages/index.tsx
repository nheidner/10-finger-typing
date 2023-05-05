import { Book } from "@/types";
import { useEffect, useState } from "react";

export default function Home() {
  const [books, setBooks] = useState<null | Book[]>(null);

  useEffect(() => {
    fetch("/api/books")
      .then((res) => res.json() as Promise<{ data: Book[] }>)
      .then((data) => setBooks(data.data));
  }, []);

  return (
    <main
      className={`flex min-h-screen flex-col items-center justify-between p-24`}
    >
      {JSON.stringify(books)}
    </main>
  );
}
