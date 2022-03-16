export { parse } from "https://deno.land/std@0.113.0/flags/mod.ts";
export {encode} from "https://deno.land/std@0.113.0/encoding/base64.ts";
export { readAll } from "https://deno.land/std@0.113.0/streams/conversion.ts";
import * as log from "https://deno.land/std@0.113.0/log/mod.ts";

await log.setup({
    handlers: {
        console: new log.handlers.ConsoleHandler("INFO")
    }
})

export const logger = log.getLogger("DUMP")