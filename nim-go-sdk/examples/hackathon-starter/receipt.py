import os
import time
import json
import requests
import cv2

try:
    # optional, but handy if you use a .env file
    from dotenv import load_dotenv
    load_dotenv()
except Exception:
    pass

TABSCANNER_API_KEY = os.getenv("TABSCANNER_API_KEY") or os.getenv("API_KEY")
if not TABSCANNER_API_KEY:
    raise SystemExit(
        "Missing API key. Set TABSCANNER_API_KEY (recommended) or API_KEY in your environment/.env"
    )

PROCESS_URL = "https://api.tabscanner.com/api/2/process"
RESULT_URL_BASE = "https://api.tabscanner.com/api/result"  # /{token}

HEADERS = {"apikey": TABSCANNER_API_KEY}


def capture_receipt_image(output_path: str = "receipt.jpg") -> str:
    """
    Opens webcam. Press SPACE to capture, ESC to quit.
    """
    cap = cv2.VideoCapture(0)
    if not cap.isOpened():
        raise RuntimeError(
            "Could not open webcam. Check permissions or try a different camera index."
        )

    print("Webcam open. Aim at the receipt.")
    print("Press SPACE to capture, ESC to cancel.")

    while True:
        ok, frame = cap.read()
        if not ok:
            continue

        cv2.imshow("Receipt Capture (SPACE=shoot, ESC=quit)", frame)
        key = cv2.waitKey(1) & 0xFF

        if key == 27:  # ESC
            cap.release()
            cv2.destroyAllWindows()
            raise SystemExit("Cancelled.")
        elif key == 32:  # SPACE
            cv2.imwrite(output_path, frame)
            cap.release()
            cv2.destroyAllWindows()
            print(f"Saved image: {output_path}")
            return output_path


def tabscanner_process(image_path: str, region: str = "gb", document_type: str = "receipt") -> str:
    """
    POST image to Tabscanner process endpoint. Returns token.
    """
    data = {
        "documentType": document_type,
        "region": region,  # optional but can improve date/number parsing
        # "defaultDateParsing": "d/m",  # optional, useful for UK receipts
    }

    with open(image_path, "rb") as f:
        files = {"file": (os.path.basename(image_path), f, "image/jpeg")}
        resp = requests.post(PROCESS_URL, headers=HEADERS, data=data, files=files, timeout=60)

    resp.raise_for_status()
    payload = resp.json()

    if not payload.get("success", True) and "token" not in payload:
        raise RuntimeError(f"Tabscanner process failed: {payload}")

    token = payload.get("token")
    if not token:
        raise RuntimeError(f"No token returned. Response: {payload}")

    return token


def tabscanner_poll_result(token: str, initial_wait_s: float = 5.0, poll_every_s: float = 1.0, timeout_s: float = 60.0):
    """
    Polls Tabscanner result endpoint until status is done (or timeout).
    """
    time.sleep(initial_wait_s)

    url = f"{RESULT_URL_BASE}/{token}"
    start = time.time()

    while True:
        resp = requests.get(url, headers=HEADERS, timeout=60)
        resp.raise_for_status()
        payload = resp.json()

        status = payload.get("status")  # done | pending | failed (per docs)
        if status == "done" and payload.get("success", True):
            return payload
        if status == "failed" or payload.get("success") is False:
            raise RuntimeError(f"Tabscanner result failed: {payload}")

        if time.time() - start > timeout_s:
            raise TimeoutError(f"Timed out waiting for result. Last payload: {payload}")

        time.sleep(poll_every_s)


def print_receipt_summary(result_payload: dict):
    """
    Prints common fields + line items if present.
    """
    result = result_payload.get("result") or {}

    establishment = result.get("establishment")
    date = result.get("date") or result.get("dateISO")
    total = result.get("total")
    sub_total = result.get("subTotal")
    tax = result.get("tax")
    currency = result.get("currency")
    payment_method = result.get("paymentMethod")
    address = result.get("address")

    print("\n=== RECEIPT SUMMARY ===")
    print(f"Merchant: {establishment}")
    print(f"Date:     {date}")
    print(f"Total:    {total} {currency or ''}".strip())
    if sub_total is not None:
        print(f"SubTotal: {sub_total} {currency or ''}".strip())
    if tax is not None:
        print(f"Tax:      {tax} {currency or ''}".strip())
    if payment_method:
        print(f"Payment:  {payment_method}")
    if address:
        print(f"Address:  {address}")

    # Line items (Tabscanner returns structured lineItems; exact key depends on receipt)
    line_items = (
        result.get("lineItems")
        or result.get("line_items")
        or result.get("items")
        or []
    )

    if isinstance(line_items, list) and line_items:
        print("\n=== LINE ITEMS ===")
        for i, item in enumerate(line_items, 1):
            desc = item.get("desc") or item.get("description") or item.get("text") or ""
            qty = item.get("qty")
            price = item.get("price")
            line_total = item.get("lineTotal") or item.get("total")
            print(f"{i:02d}. {desc}".strip())
            bits = []
            if qty is not None:
                bits.append(f"qty={qty}")
            if price is not None:
                bits.append(f"price={price}")
            if line_total is not None:
                bits.append(f"lineTotal={line_total}")
            if bits:
                print("    " + ", ".join(bits))
    else:
        print("\n(No line items detected, or receipt didn’t include itemised rows.)")

    # If you want the full raw JSON, uncomment:
    # print("\n=== RAW JSON ===")
    # print(json.dumps(result_payload, indent=2))


def main():
    image_path = capture_receipt_image("receipt.jpg")
    print("Uploading to Tabscanner...")
    token = tabscanner_process(image_path, region="gb", document_type="receipt")
    print(f"Token: {token}")

    print("Waiting for result...")
    result_payload = tabscanner_poll_result(token, initial_wait_s=5, poll_every_s=1, timeout_s=90)

    print_receipt_summary(result_payload)


if __name__ == "__main__":
    main()
