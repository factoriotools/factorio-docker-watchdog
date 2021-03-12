import hashlib
import re
import json
import requests
from requests import HTTPError


def get_max_prefix(prefixes):
    to_return = None
    candidate_major = None
    candidate_minor = None
    for prefix in prefixes:
        major, minor = list(map(int, prefix.split(".")))
        if to_return is None or major > candidate_major or (major == candidate_major and minor > candidate_minor):
            candidate_major = major
            candidate_minor = minor
            to_return = prefix
    return to_return


def write_buildinfo(buildinfo):
    with open("buildinfo.json", "w") as buildinfo_file:
        json.dump(buildinfo, buildinfo_file, indent=2, sort_keys=True)


def modify_readme(buildinfo):
    generated_content = "\n"
    for version in buildinfo:
        generated_content += "* " + ", ".join(map(lambda x: f"`{x}`", buildinfo[version]["tags"])) + "\n"
    with open("README.md", "r") as readme_file:
        readme = readme_file.read()
    readme = re.sub(r"(<!-- start autogeneration tags -->).*(<!-- end autogeneration tags -->)",
                    fr"\1{generated_content}\2", readme, flags=re.MULTILINE | re.DOTALL)
    with open("README.md", "w") as readme_file:
        readme_file.write(readme)


def has_diff(new_buildinfo):
    with open("buildinfo.json") as buildinfo_file:
        old_buildinfo = json.load(buildinfo_file)
    if len(old_buildinfo) != len(new_buildinfo):
        return True
    for version in new_buildinfo:
        if version not in old_buildinfo or old_buildinfo[version]["tags"] != new_buildinfo[version]["tags"]:
            return True
    return False


def get_sha1_hash(version):
    sha1_hash = hashlib.sha1()
    try:
        with requests.get(f"https://www.factorio.com/get-download/{version}/headless/linux64", stream=True) as r:
            r.raise_for_status()
            for chunk in r.iter_content(chunk_size=8192):
                sha1_hash.update(chunk)
    except HTTPError as e:
        print(f"Error {e.response.status_code} on Download of version {version}")
        exit(1)
    return sha1_hash.hexdigest()


def loop():
    result = requests.get("https://www.factorio.com/updater/get-available-versions?apiVersion=2")
    json_response = result.json()
    stable_version = None
    highest_version_per_prefix = {}
    # Get latest patches + stable version
    for version in json_response["core-linux_headless64"]:
        if "stable" in version:
            stable_version = version["stable"]
        if "to" in version:
            prefix, patch = version["to"].rsplit(".", 1)
            patch = int(patch)
            if prefix not in highest_version_per_prefix or highest_version_per_prefix[prefix] < patch:
                highest_version_per_prefix[prefix] = patch

    # Generate buildinfo standard tags
    buildinfo = {}
    for prefix, patch in highest_version_per_prefix.items():
        buildinfo[f"{prefix}.{patch}"] = {
            "tags": [f"{prefix}", f"{prefix}.{patch}"]
        }

    # Add "latest" and "stable" tag
    max_prefix = get_max_prefix(highest_version_per_prefix.keys())
    max_patch = highest_version_per_prefix[max_prefix]
    buildinfo[f"{max_prefix}.{max_patch}"]["tags"].append("latest")
    if stable_version:
        buildinfo.setdefault(stable_version, {"tags": []})
        buildinfo[stable_version]["tags"].append("stable")

    if not has_diff(buildinfo):
        return

    # Add sha1 for each version
    for version, infos in buildinfo.items():
        infos["sha1"] = get_sha1_hash(version)

    # Write buildinfo to file
    write_buildinfo(buildinfo)

    # Modify readme
    modify_readme(buildinfo)


if __name__ == '__main__':
    loop()
