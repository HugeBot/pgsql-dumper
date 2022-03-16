// Requires: --allow-env --allow-write --allow-read --allow-net

import { encode, logger, parse, readAll } from "./deps.ts";

(async () => {
    logger.info("Starting Postgres Dumper...\n")

    const tmp = Deno.env.get("TMPDIR") || Deno.env.get("TMP") || Deno.env.get("TEMP") || "/tmp";

    const aliases = {
        dbName: "N",
        owner: "O",
        repo: "R",
        token: "T"
    }

    const date = new Date()

    const args = parse(Deno.args, {
        alias: aliases
    })

    const dbName = args.dbName
    const owner = args.owner
    const repo = args.repo
    const token = args.token

    if (!dbName) throw new Error("Database Name must be provided (use -N <name>)")
    if (!owner) throw new Error("GitHub Owner must be provided (use -O <owner>)")
    if (!repo) throw new Error("GitHub Repo must be provided (use -R <repo>)")
    if (!token) throw new Error("GitHub Token must be provided (use -T <token>)")

    const fileName = `dump-${dbName}-${date.toTimeString()}.bck`
    
    logger.info("Creating dump process...")
    const p = Deno.run({
        cmd: ["pg_dump", "-Z5", "-Fc", dbName, ">", tmp+"/"+fileName],
        stdout: "piped",
        stderr: "piped",
        stdin: "piped"
    })
    
    const code = await p.status()

    if (!code.success) {
        throw new Error("Non-successful status code from process: " + code.code)
    }

    logger.info(`Dump created on ${tmp}/${fileName}`)

    logger.info(`B64 encoding file ${tmp}/${fileName}...`)
    const file = await Deno.open(tmp + "/" + fileName)
    const encoded = encode(await readAll(file))
    logger.info(`${tmp}/${fileName} encoded!`)

    const data = {
        "message": `Upload backup from ${dbName} at ${date.toTimeString()}`,
        "content": encoded
    }

    logger.info(`Uploading file to GitHub...`)
    const res = await fetch(`https://api.github.com/repos/${owner}/${repo}/contents/${fileName}`, {
        headers: {
            "Authorization": `token ${token}`
        },
        body: JSON.stringify(data)
    })

    if (res.status !== 201) {
        throw new Error("Non-successful status code uploading backup file: " + res.statusText)
    }

    const json = await res.json()

    logger.info(`File upload successfully to GitHub, response:`)
    logger.info(json)
    logger.info("\n")
    logger.info("Done!\n")


})()