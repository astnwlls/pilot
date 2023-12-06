# mock_step.py

import sys

def main():
    step_name = sys.argv[1] if len(sys.argv) > 1 else "Unnamed Step"
    print(f"Executing mock step: {step_name}")

if __name__ == "__main__":
    main()
