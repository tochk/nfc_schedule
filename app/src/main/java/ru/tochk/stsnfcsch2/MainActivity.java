package ru.tochk.stsnfcsch2;

import android.content.Intent;
import android.nfc.NfcAdapter;
import android.nfc.Tag;
import android.os.Bundle;
import android.os.StrictMode;
import android.support.v7.app.AppCompatActivity;
import android.view.View;
import android.widget.TextView;
import android.widget.Toast;

import com.blogspot.android_er.androidnfctagdiscovered.R;

import java.io.IOException;
import java.util.Objects;

public class MainActivity extends AppCompatActivity {

    private NfcAdapter nfcAdapter;
    public TextView textViewInfo;
    public TextView userViewInfo;
    public TextView positionViewInfo;
    public String tempTagFullClean;
    public User user = new User();
    private String lastTag = "";

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);
        textViewInfo = (TextView) findViewById(R.id.info);
        nfcAdapter = NfcAdapter.getDefaultAdapter(this);
        if (nfcAdapter == null) {
            Toast.makeText(this,
                    "NFC NOT supported on this device!",
                    Toast.LENGTH_LONG).show();
            finish();
        } else if (!nfcAdapter.isEnabled()) {
            Toast.makeText(this,
                    "NFC NOT Enabled!",
                    Toast.LENGTH_LONG).show();
            finish();
        }
    }

    @Override
    protected void onResume() {
        super.onResume();

        Intent intent = getIntent();
        String action = intent.getAction();

        if (NfcAdapter.ACTION_TAG_DISCOVERED.equals(action)) {
            Tag tag = intent.getParcelableExtra(NfcAdapter.EXTRA_TAG);
            if (tag == null) {
                textViewInfo.setText("tag == null");
            } else {
                String tagInfo = "";

                byte[] tagId = tag.getId();
                String tempTag;
                String tempTagFull = "";
                tempTagFullClean = "";
                for (byte aTagId : tagId) {
                    tempTag = Integer.toHexString(aTagId & 0xFF);
                    if (tempTag.length() == 1) {
                        tempTag = "0" + tempTag;
                    }
                    if (tempTag.length() == 0) {
                        tempTag = "00";
                    }
                    tempTagFull += tempTag + ":";
                    this.tempTagFullClean += tempTag;


                }

                System.out.println(lastTag);
                System.out.println(tempTagFullClean);

                if (this.tempTagFullClean.equals(this.lastTag)) {
                    return;
                } else {
                    this.lastTag = this.tempTagFullClean;
                }

                tagInfo += "Идентификатор карты\n" + tempTagFull.substring(0, tempTagFull.length() - 1) + "\n\n";


                textViewInfo.setText(tagInfo);
                Thread thread = new Thread(new Runnable() {
                    public void run() {
                        try {
                            user.getUserInfo(tempTagFullClean);
                        } catch (Exception e) {
                            e.printStackTrace();
                        }
                    }
                });

                thread.start();

                try {
                    thread.join();
                } catch (InterruptedException e) {
                    e.printStackTrace();
                }

                userViewInfo = (TextView) findViewById(R.id.user);
                positionViewInfo = (TextView) findViewById(R.id.position);
                userViewInfo.setText(user.fullName);
                positionViewInfo.setText(user.position);
                switch (user.isStart) {
                    case "true":
                        positionViewInfo.getRootView().setBackgroundColor(getResources().getColor(android.R.color.holo_green_light));
                        break;
                    case "false":
                        positionViewInfo.getRootView().setBackgroundColor(getResources().getColor(android.R.color.holo_orange_light));
                        break;
                    case "notfound":
                        positionViewInfo.getRootView().setBackgroundColor(getResources().getColor(android.R.color.holo_red_light));
                        break;
                }


            }
        }
    }

    @Override
    protected void onPause() {
        super.onPause();

        nfcAdapter = NfcAdapter.getDefaultAdapter(this);
    }
}