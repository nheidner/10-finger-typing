export const debounce = <T>(fn: (...args: any[]) => any, interval: number) => {
  let timeout: NodeJS.Timeout;

  return (...args: any[]): any => {
    return new Promise((resolve, reject) => {
      clearTimeout(timeout);

      timeout = setTimeout(async () => {
        try {
          const result = await fn(...args);
          resolve(result);
        } catch (error) {
          reject(error);
        }
      }, interval);
    });
  };
};
