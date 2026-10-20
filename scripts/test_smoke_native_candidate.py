"""Reject candidate identity drift and unsafe archive contents before launch."""
import hashlib
import importlib.util
import io
from pathlib import Path
import tarfile
import tempfile
import unittest
import zipfile

spec = importlib.util.spec_from_file_location("native", Path(__file__).with_name("smoke-native-candidate.py"))
native = importlib.util.module_from_spec(spec)
spec.loader.exec_module(native)


class CandidateIdentity(unittest.TestCase):
    def test_sha_and_checksum_must_both_match(self):
        with tempfile.TemporaryDirectory() as scratch:
            root = Path(scratch)
            archive = root / "candidate.zip"
            archive.write_bytes(b"candidate")
            checksums = root / "checksums.txt"
            checksums.write_text(hashlib.sha256(b"candidate").hexdigest() + "  candidate.zip\n")
            sha = "a" * 40
            native.validate(archive, checksums, {"commit": sha}, sha)
            with self.assertRaisesRegex(RuntimeError, "full workflow SHA"):
                native.validate(archive, checksums, {"commit": sha[:7]}, sha)
            archive.write_bytes(b"different candidate")
            with self.assertRaisesRegex(RuntimeError, "digest"):
                native.validate(archive, checksums, {"commit": sha}, sha)

    def test_zip_and_tar_require_only_release_members(self):
        with tempfile.TemporaryDirectory() as scratch:
            root = Path(scratch)
            archive = root / "candidate.zip"
            with zipfile.ZipFile(archive, "w") as bundle:
                for name in ("bitbottle.exe", "README.md", "LICENSE"):
                    bundle.writestr(name, b"fixture")
            self.assertEqual(native.unpack(archive, root).read_bytes(), b"fixture")
            with zipfile.ZipFile(archive, "a") as bundle:
                bundle.writestr("../escape", b"bad")
            with self.assertRaisesRegex(RuntimeError, "ZIP content"):
                native.unpack(archive, root)
            archive = root / "candidate.tar.gz"
            with tarfile.open(archive, "w:gz") as bundle:
                for name in ("bitbottle", "README.md", "LICENSE"):
                    info = tarfile.TarInfo(name)
                    info.size = 7
                    bundle.addfile(info, io.BytesIO(b"fixture"))
            self.assertEqual(native.unpack(archive, root).read_bytes(), b"fixture")


if __name__ == "__main__":
    unittest.main()
