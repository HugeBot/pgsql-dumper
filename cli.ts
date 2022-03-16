// Requires: --allow-env --allow-write --allow-read --allow-net

import { encode, parse, readAll } from "./deps.ts";

(async () => {
    console.info("Starting Postgres Dumper...\n")

    const tmp = Deno.env.get("TMPDIR") || Deno.env.get("TMP") || Deno.env.get("TEMP") || "/tmp";

    console.info("Temp dir is " + tmp)

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

    const fileName = `dump-${dbName}-${date.toISOString()}.bck`
    
    console.info(`Creating dump process, file will be saved on ${tmp}/${fileName}...`)

    const command = ["pg_dump", "-Z5", "-Fc", `${dbName}`, "-f", `${tmp}/${fileName}`]

    console.info("Executing process with command " + command.join(" "))

    const p = Deno.run({
        cmd: command,
        stdout: "inherit",
        stderr: "inherit"
    })
    
    const code = await p.status()

    if (!code.success) {
        throw new Error("Non-successful status code from process: " + code.code)
    }

    console.info(`Dump created on ${tmp}/${fileName}`)

    console.info(`B64 encoding file ${tmp}/${fileName}...`)
    const file = await Deno.open(tmp + "/" + fileName)
    const encoded = encode(await readAll(file))
    console.info(`${tmp}/${fileName} encoded!`)

    const data = {
        "message": `Upload backup from ${dbName} at ${date.toISOString()}`,
        "content": encoded
    }

    console.info(`Uploading file to GitHub...`)
    const res = await fetch(`https://api.github.com/repos/${owner}/${repo}/contents/${fileName}`, {
        method: "PUT",
        headers: {
            "Authorization": `token ${token}`
        },
        body: JSON.stringify(data)
    })

    if (res.status !== 201) {
        throw new Error("Non-successful status code uploading backup file: " + res.statusText)
    }

    const json = await res.json()

    console.info(`File upload successfully to GitHub, response:`)
    console.info(json)
    console.info("\n")
    console.info("Done!\n")


})()