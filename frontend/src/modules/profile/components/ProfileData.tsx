import type { User } from "@/types";

export const ProfileData = ({ user }: { user?: User }) => {
  if (!user) {
    return null;
  }

  const data = [
    { name: "Username", value: user.username },
    { name: "Email", value: user.email },
    { name: "Full name", value: `${user.firstName} ${user.lastName}` },
  ];

  return (
    <div className="flex-1">
      <div className="px-4 sm:px-0">
        <h3 className="text-base font-semibold leading-7 text-gray-900">
          Personal Information
        </h3>
        {/* <p className="mt-1 max-w-2xl text-sm leading-6 text-gray-500">
          Personal details and application.
        </p> */}
      </div>
      <div className="mt-6 border-t border-gray-100">
        {/* <dl className="divide-y divide-gray-100"> */}
        <dl className="">
          {data.map((item) => {
            return (
              <div
                key={item.name}
                className="px-4 py-6 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-0"
              >
                <dt className="text-sm font-medium leading-6 text-gray-900">
                  {item.name}
                </dt>
                <dd className="mt-1 text-sm leading-6 text-gray-700 sm:col-span-2 sm:mt-0">
                  {item.value}
                </dd>
              </div>
            );
          })}
        </dl>
      </div>
    </div>
  );
};
