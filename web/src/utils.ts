import { Buffer } from "buffer";

const alphaOnly = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz';

export function genString(n: number):string {
  let result = '';
  for (let i = 0; i < n; i++) {
    result += alphaOnly.charAt(Math.floor(Math.random() * alphaOnly.length));
  }
  return result;
}


export const encodeBase64 = (str:string) => {
  const b = Buffer.from(str);
  return b.toString("base64");
};

export const decodeBase64 = (b64str : string) => {
  const b = Buffer.from(b64str, "base64");
  return b.toString();
};
